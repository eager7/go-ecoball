#!/usr/bin/python

from jinja2 import Template

template = Template(''' 
{
	"PublicKey":{{node.PublicKey}},
	"Address":{{node.Address}},
	"Port":{{node.Port}},
	"Size":{{node.Size}},
	{% for group in shards|groupby('role') %}
	"{{ group.grouper }}":[
		{% for member in group.list %}
		{
			"PublicKey":{{ member.PublickKey }},
			"Address":{{ member.Address }},
			"Port":{{ member.Port }}
		},
		{% endfor %}
	]
	{% endfor %}
}
''')

shards = []
shards.append({
	"role":"Committee",
	"PublickKey":"3322",
	"Address":"0.1.1.2",
	"Port":"100"
	})
shards.append({
	"role":"Shard",
	"PublickKey":"1212",
	"Address":"1.2.1.2", 
	"Port":"200"})
shards.append({
	"role":"Shard",
	"PublickKey":"0101",
	"Address":"0.1.0.1",
	"Port":"201"})

self = {
	"PublicKey":"0101",
	"Address":"0.1.0.1",
	"Port":"201",
	"Size":"1"}

print(template.render(node=self, shards=shards))
