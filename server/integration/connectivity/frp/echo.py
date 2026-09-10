import socket
import threading


def tcp_server():
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind(("0.0.0.0", 52000))
    server.listen(16)
    while True:
        conn, _ = server.accept()
        with conn:
            payload = conn.recv(1024)
            conn.sendall(b"frp-tcp-ok:" + payload)


def udp_server():
    server = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    server.bind(("0.0.0.0", 52000))
    server.settimeout(1)
    while True:
        try:
            payload, address = server.recvfrom(1024)
        except socket.timeout:
            continue
        server.sendto(b"frp-udp-ok:" + payload, address)


threading.Thread(target=tcp_server, daemon=True).start()
udp_server()
