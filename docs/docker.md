# Docker (WIP)

## Running:
### Chromium container:
```
sudo docker run -p 80:8080 -p 59000-59100:59000-59100/udp -e NEKO_PASSWORD='secret' -e NEKO_PASSWORD_ADMIN='secret' --cap-add SYS_ADMIN --shm-size=1gb nurdism/neko:chromium
```
*Note:* `--cap-add SYS_ADMIN` & `--shm-size=1gb` is required for chromium to run properly

----
### Firefox container:
```
sudo docker run -p 8080:8080 -p 59000-59100:59000-59100/udp -e NEKO_PASSWORD='secret' -e NEKO_PASSWORD_ADMIN='secret' --shm-size=1gb nurdism/neko:firefox 
```
*Note:* `--shm-size=1gb` is required for firefox, tabs will crash otherwise