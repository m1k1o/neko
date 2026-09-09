import socket
import sys
import time


host = sys.argv[1]
port = int(sys.argv[2])
payload = b"m1-connectivity"


def retry(operation):
    last_error = None
    for _ in range(30):
        try:
            return operation()
        except OSError as error:
            last_error = error
            time.sleep(1)
    raise RuntimeError(f"FRP endpoint {host}:{port} is unreachable: {last_error}")


def tcp_check():
    def operation():
        with socket.create_connection((host, port), timeout=2) as connection:
            connection.sendall(payload)
            response = connection.recv(1024)
            if response != b"frp-tcp-ok:" + payload:
                raise RuntimeError(f"unexpected TCP response: {response!r}")

    retry(operation)
    print(f"TCP {host}:{port}: forwarded")


def udp_check():
    def operation():
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as connection:
            connection.settimeout(2)
            connection.sendto(payload, (host, port))
            response, _ = connection.recvfrom(1024)
            if response != b"frp-udp-ok:" + payload:
                raise RuntimeError(f"unexpected UDP response: {response!r}")

    retry(operation)
    print(f"UDP {host}:{port}: forwarded")


tcp_check()
udp_check()
