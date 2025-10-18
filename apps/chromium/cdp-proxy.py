#!/usr/bin/env python3
"""
Simple TCP proxy to forward CDP connections from 0.0.0.0:9223 to 127.0.0.1:9222
This works around Chromium's limitation of only binding CDP to localhost.
"""
import socket
import select
import sys

def proxy_connection(client_sock, target_host='127.0.0.1', target_port=9222):
    """Forward data between client and target"""
    try:
        target_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        target_sock.connect((target_host, target_port))

        sockets = [client_sock, target_sock]

        while True:
            readable, _, exceptional = select.select(sockets, [], sockets)

            if exceptional:
                break

            for sock in readable:
                data = sock.recv(8192)
                if not data:
                    return

                if sock is client_sock:
                    target_sock.sendall(data)
                else:
                    client_sock.sendall(data)

    except Exception as e:
        print(f"Connection error: {e}", file=sys.stderr)
    finally:
        client_sock.close()
        target_sock.close()

def main():
    listen_host = '0.0.0.0'
    listen_port = 9223

    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind((listen_host, listen_port))
    server.listen(5)

    print(f"CDP proxy listening on {listen_host}:{listen_port}, forwarding to 127.0.0.1:9222")

    try:
        while True:
            client_sock, addr = server.accept()
            print(f"New connection from {addr}")

            # Handle in the same thread for simplicity (could use threading for multiple connections)
            import threading
            threading.Thread(target=proxy_connection, args=(client_sock,), daemon=True).start()
    except KeyboardInterrupt:
        print("\nShutting down...")
    finally:
        server.close()

if __name__ == '__main__':
    main()
