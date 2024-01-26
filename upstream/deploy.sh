#!/usr/bin/bash

rsync -azP -e "ssh -i /home/thuan/keys/nguyenphuochoangthuan02.pem" . ubuntu@168.138.8.216:/var/www/html/cave/neko-cave/upstream/
