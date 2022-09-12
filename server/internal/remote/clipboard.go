package remote

import "m1k1o/neko/internal/remote/clipboard"

func (manager *RemoteManager) ReadClipboard() string {
	return clipboard.Read()
}

func (manager *RemoteManager) WriteClipboard(data string) {
	clipboard.Write(data)
}
