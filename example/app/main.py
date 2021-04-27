from http.server import BaseHTTPRequestHandler, HTTPServer
from socketserver import ThreadingMixIn

class ThreadedHTTPServer(ThreadingMixIn, HTTPServer):
    daemon_thread = True

class DefaultHandler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def do_GET(self):
        path, _, _ = self.path.partition('?')

        if path == "/":
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header('Connection', 'close')
            self.end_headers()
            self.wfile.write(bytes("""{"sample":true}\r\n"""))
            self.wfile.flush()
        elif path.startswith("/healthcheck"):
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.send_header('Connection', 'close')
            self.end_headers()
            self.wfile.write(bytes("WORKING\r\n"))
            self.wfile.flush()
        self.wfile.close()
        return

if __name__ == "__main__":
    server = ThreadedHTTPServer(("", 8888), DefaultHandler)
    server.serve_forever()