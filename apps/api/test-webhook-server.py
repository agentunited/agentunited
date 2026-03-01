#!/usr/bin/env python3
"""
Simple test webhook server to capture and display webhook requests from Agent United.
"""
import json
import time
from http.server import HTTPServer, BaseHTTPRequestHandler

class WebhookHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        # Get content length
        content_length = int(self.headers.get('content-length', 0))
        
        # Read the body
        body = self.rfile.read(content_length)
        
        # Parse JSON payload
        try:
            payload = json.loads(body.decode('utf-8'))
        except:
            payload = body.decode('utf-8')
        
        # Display the webhook
        print("=" * 60)
        print(f"🪝 WEBHOOK RECEIVED at {time.strftime('%Y-%m-%d %H:%M:%S')}")
        print("=" * 60)
        print(f"Method: {self.command}")
        print(f"Path: {self.path}")
        print("\nHeaders:")
        for header, value in self.headers.items():
            print(f"  {header}: {value}")
        print(f"\nPayload:")
        print(json.dumps(payload, indent=2))
        print("=" * 60)
        print()
        
        # Send success response
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        response = json.dumps({"status": "ok", "received": True})
        self.wfile.write(response.encode('utf-8'))
    
    def log_message(self, format, *args):
        # Disable default HTTP logging to keep output clean
        pass

if __name__ == "__main__":
    port = 8081
    server = HTTPServer(('localhost', port), WebhookHandler)
    print(f"🚀 Test webhook server running on http://localhost:{port}")
    print("   Waiting for webhooks from Agent United...")
    print("   Press Ctrl+C to stop")
    print()
    
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n👋 Webhook server stopped")
        server.shutdown()