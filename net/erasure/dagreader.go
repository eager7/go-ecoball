package erasure

import (
	"context"
	"errors"
	"fmt"
	"io"

	mdag "gx/ipfs/QmRy4Qk9hbgFX9NGJRm8rBThrA8PZhNCitMgeRYyZ67s59/go-merkledag"
	ft "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs"
	ftpb "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs/pb"

	cid "gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
	uio "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs/io"
)

// EraSureDagReader provides a way to easily read the data contained in a dag.
type EraSureDagReader struct {
	serv ipld.NodeGetter

	// UnixFS file (it should be of type `Data_File` or `Data_Raw` only).
	file *ft.FSNode

	// the current data buffer to be read from
	// will either be a bytes.Reader or a child DagReader
	buf uio.ReadSeekCloser

	// NodePromises for each of 'nodes' child links
	promises []*ipld.NodePromise

	// the cid of each child of the current node
	links []*cid.Cid

	// the index of the child link currently being read from
	linkPosition int

	// current offset for the read head within the 'file'
	offset int64

	// Our context
	ctx context.Context

	// context cancel for children
	cancel func()
}

//var _ DagReader = (*EraSureDagReader)(nil)

// NewPBFileReader constructs a new PBFileReader.
func NewErasureFileReader(ctx context.Context, n *mdag.ProtoNode, file *ft.FSNode, serv ipld.NodeGetter) *EraSureDagReader {
	fctx, cancel := context.WithCancel(ctx)
	curLinks := getLinkCidsExcludeEra(n)
	return &EraSureDagReader{
		serv:     serv,
		buf:      uio.NewBufDagReader(file.Data()),
		promises: make([]*ipld.NodePromise, len(curLinks)),
		links:    curLinks,
		ctx:      fctx,
		cancel:   cancel,
		file:     file,
	}
}

const preloadSize = 10

func (dr *EraSureDagReader) preload(ctx context.Context, beg int) {
	end := beg + preloadSize
	if end >= len(dr.links) {
		end = len(dr.links)
	}

	copy(dr.promises[beg:], ipld.GetNodes(ctx, dr.serv, dr.links[beg:end]))
}

// precalcNextBuf follows the next link in line and loads it from the
// DAGService, setting the next buffer to read from
func (dr *EraSureDagReader) precalcNextBuf(ctx context.Context) error {
	if dr.buf != nil {
		dr.buf.Close() // Just to make sure
		dr.buf = nil
	}

	if dr.linkPosition >= len(dr.promises) {
		return io.EOF
	}

	// If we drop to <= preloadSize/2 preloading nodes, preload the next 10.
	for i := dr.linkPosition; i < dr.linkPosition+preloadSize/2 && i < len(dr.promises); i++ {
		// TODO: check if canceled.
		if dr.promises[i] == nil {
			fmt.Printf("precalcNextBuf preload-----\n")
			dr.preload(ctx, i)
			break
		}
	}

	nxt, err := dr.promises[dr.linkPosition].Get(ctx)
	dr.promises[dr.linkPosition] = nil
	switch err {
	case nil:
	case context.DeadlineExceeded, context.Canceled:
		err = ctx.Err()
		if err != nil {
			return ctx.Err()
		}
		// In this case, the context used to *preload* the node has been canceled.
		// We need to retry the load with our context and we might as
		// well preload some extra nodes while we're at it.
		//
		// Note: When using `Read`, this code will never execute as
		// `Read` will use the global context. It only runs if the user
		// explicitly reads with a custom context (e.g., by calling
		// `CtxReadFull`).
		fmt.Printf("precalcNextBuf preload+++\n")
		dr.preload(ctx, dr.linkPosition)
		nxt, err = dr.promises[dr.linkPosition].Get(ctx)
		dr.promises[dr.linkPosition] = nil
		if err != nil {
			return err
		}
	default:
		return err
	}

	dr.linkPosition++

	//for test
	//if dr.linkPosition == 2 {
	//	return ErrUnkownNodeType
	//}

	return dr.loadBufNode(nxt)
}

func (dr *EraSureDagReader) loadBufNode(node ipld.Node) error {
	fmt.Printf("loadBufNode, cid: %s\n", node.Cid().String())
	switch node := node.(type) {
	case *mdag.ProtoNode:
		fsNode, err := ft.FSNodeFromBytes(node.Data())
		if err != nil {
			return fmt.Errorf("incorrectly formatted protobuf: %s", err)
		}

		switch fsNode.Type() {
		case ftpb.Data_File:
			dr.buf = uio.NewPBFileReader(dr.ctx, node, fsNode, dr.serv)
			return nil
		case ftpb.Data_Raw:
			dr.buf = uio.NewBufDagReader(fsNode.Data())
			return nil
		default:
			return fmt.Errorf("found %s node in unexpected place", fsNode.Type().String())
		}
	case *mdag.RawNode:
		dr.buf = uio.NewBufDagReader(node.RawData())
		return nil
	default:
		return uio.ErrUnkownNodeType
	}
}

func getLinkCidsExcludeEra(n ipld.Node) []*cid.Cid {
	links := n.Links()
	out := make([]*cid.Cid, 0, len(links))
	for _, l := range links {
		if l.Name == "erasure" {
			continue
		}
		out = append(out, l.Cid)
	}
	return out
}

func getEraLink(n ipld.Node) *cid.Cid {
	links := n.Links()
	for _, l := range links {
		if l.Name == "erasure" {
			return l.Cid
		}
	}
	return nil
}

// Size return the total length of the data from the DAG structured file.
func (dr *EraSureDagReader) Size() uint64 {
	return dr.file.FileSize()
}

// Read reads data from the DAG structured file
func (dr *EraSureDagReader) Read(b []byte) (int, error) {
	return dr.CtxReadFull(dr.ctx, b)
}

// CtxReadFull reads data from the DAG structured file
func (dr *EraSureDagReader) CtxReadFull(ctx context.Context, b []byte) (int, error) {
	if dr.buf == nil {
		if err := dr.precalcNextBuf(ctx); err != nil {
			return 0, err
		}
	}

	// If no cached buffer, load one
	total := 0
	for {
		// Attempt to fill bytes from cached buffer
		n, err := io.ReadFull(dr.buf, b[total:])
		total += n
		dr.offset += int64(n)
		switch err {
		// io.EOF will happen is dr.buf had noting more to read (n == 0)
		case io.EOF, io.ErrUnexpectedEOF:
			// do nothing
		case nil:
			return total, nil
		default:
			return total, err
		}

		// if we are not done with the output buffer load next block
		err = dr.precalcNextBuf(ctx)
		if err != nil {
			return total, err
		}
		fmt.Printf("CtxReadFull total: %d\n", total)
	}
}

// WriteTo writes to the given writer.
func (dr *EraSureDagReader) WriteTo(w io.Writer) (int64, error) {
	fmt.Printf("1 start Writeto--->\n")
	if dr.buf == nil {
		if err := dr.precalcNextBuf(dr.ctx); err != nil {
			return 0, err
		}
	}

	// If no cached buffer, load one
	total := int64(0)
	for {
		fmt.Printf("2 start total: %d\n", total)
		// Attempt to write bytes from cached buffer
		n, err := dr.buf.WriteTo(w)
		fmt.Printf("3 start n: %d\n", n)
		total += n
		dr.offset += n
		if err != nil {
			if err != io.EOF {
				return total, err
			}
		}
		fmt.Printf("4 WriteTo, total: %d\n", total)
		// Otherwise, load up the next block
		err = dr.precalcNextBuf(dr.ctx)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("5 ****WriteTo end---, total: %d\n", total)
				return total, nil
			}
			return total, err
		}

		fmt.Printf("6 WriteTo, total: %d\n", total)
	}
}

// Close closes the reader.
func (dr *EraSureDagReader) Close() error {
	dr.cancel()
	return nil
}

// Seek implements io.Seeker, and will seek to a given offset in the file
// interface matches standard unix seek
// TODO: check if we can do relative seeks, to reduce the amount of dagreader
// recreations that need to happen.
func (dr *EraSureDagReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		if offset < 0 {
			return -1, errors.New("invalid offset")
		}
		if offset == dr.offset {
			return offset, nil
		}

		// left represents the number of bytes remaining to seek to (from beginning)
		left := offset
		if int64(len(dr.file.Data())) >= offset {
			// Close current buf to close potential child dagreader
			if dr.buf != nil {
				dr.buf.Close()
			}
			dr.buf = uio.NewBufDagReader(dr.file.Data()[offset:])

			// start reading links from the beginning
			dr.linkPosition = 0
			dr.offset = offset
			return offset, nil
		}

		// skip past root block data
		left -= int64(len(dr.file.Data()))

		// iterate through links and find where we need to be
		for i := 0; i < dr.file.NumChildren(); i++ {
			if dr.file.BlockSize(i) > uint64(left) {
				dr.linkPosition = i
				break
			} else {
				left -= int64(dr.file.BlockSize(i))
			}
		}

		// start sub-block request
		err := dr.precalcNextBuf(dr.ctx)
		if err != nil {
			return 0, err
		}

		// set proper offset within child readseeker
		n, err := dr.buf.Seek(left, io.SeekStart)
		if err != nil {
			return -1, err
		}

		// sanity
		left -= n
		if left != 0 {
			return -1, errors.New("failed to seek properly")
		}
		dr.offset = offset
		return offset, nil
	case io.SeekCurrent:
		// TODO: be smarter here
		if offset == 0 {
			return dr.offset, nil
		}

		noffset := dr.offset + offset
		return dr.Seek(noffset, io.SeekStart)
	case io.SeekEnd:
		noffset := int64(dr.file.FileSize()) - offset
		n, err := dr.Seek(noffset, io.SeekStart)

		// Return negative number if we can't figure out the file size. Using io.EOF
		// for this seems to be good(-enough) solution as it's only returned by
		// precalcNextBuf when we step out of file range.
		// This is needed for gateway to function properly
		if err == io.EOF && dr.file.Type() == ftpb.Data_File {
			return -1, nil
		}
		return n, err
	default:
		return 0, errors.New("invalid whence")
	}
}


func (dr *EraSureDagReader) ErasureWriteTo([]byte)  error {
	return nil
}

func (dr *EraSureDagReader) getErasureDataSize(node ipld.Node) (uint64, error) {
	links := node.Links()
	for _, link := range links {
		if link.Name == "erasure" {
			return link.Size, nil
		}
	}
	err := errors.New("unErasuring node")
	return 0, err
}

func (dr *EraSureDagReader) getErasureLink(node ipld.Node) *ipld.Link {
	links := node.Links()
	for _, link := range links {
		if link.Name == "erasure" {
			return link
		}
	}

	return nil
}
