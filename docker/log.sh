while true; do reset && docker logs -f --tail=1000 ecoball_$1;done
