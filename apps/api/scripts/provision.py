#!/usr/bin/env python3
"""
Agent United Bootstrap Provisioning Script

This script demonstrates how an AI agent can provision a fresh Agent United instance
programmatically. It creates the primary agent, additional worker agents, human users,
and a default communication channel in a single atomic API call.

Usage:
    python provision.py [--url URL] [--save-credentials FILE]
    
Examples:
    # Provision local instance
    python provision.py
    
    # Provision remote instance
    python provision.py --url https://your-domain.com
    
    # Save credentials to custom file
    python provision.py --save-credentials ./my-credentials.json
"""

import argparse
import atexit
import json
import re
import secrets
import signal
import subprocess
import sys
import threading
import time
import urllib.parse
from typing import Dict, Any, Optional

try:
    import requests
except ImportError:
    print("Error: 'requests' library not found. Install with: pip install requests")
    sys.exit(1)


class LocaltunnelManager:
    """Manages a localtunnel subprocess for exposing local services."""
    
    def __init__(self, port: int, subdomain: Optional[str] = None):
        self.port = port
        self.subdomain = subdomain
        self.process: Optional[subprocess.Popen] = None
        self.tunnel_url: Optional[str] = None
        self._stop_event = threading.Event()
        
    def start(self) -> Optional[str]:
        """Start the localtunnel process and return the public URL."""
        if self.process is not None:
            print("⚠️  Tunnel already running")
            return self.tunnel_url
            
        print(f"🚇 Starting localtunnel for port {self.port}...")
        
        # Build localtunnel command
        cmd = ['npx', 'localtunnel', '--port', str(self.port)]
        if self.subdomain:
            cmd.extend(['--subdomain', self.subdomain])
            
        try:
            self.process = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                bufsize=1,
                universal_newlines=True
            )
            
            # Start thread to monitor output and extract URL
            monitor_thread = threading.Thread(target=self._monitor_output, daemon=True)
            monitor_thread.start()
            
            # Wait for URL to be available (timeout after 30 seconds)
            start_time = time.time()
            while self.tunnel_url is None and time.time() - start_time < 30:
                if self.process.poll() is not None:
                    # Process has terminated
                    stderr_output = self.process.stderr.read() if self.process.stderr else ""
                    print(f"❌ Localtunnel process failed: {stderr_output}")
                    return None
                time.sleep(0.1)
                
            if self.tunnel_url:
                print(f"✅ Tunnel URL: {self.tunnel_url}")
                return self.tunnel_url
            else:
                print("❌ Failed to get tunnel URL within timeout")
                self.stop()
                return None
                
        except FileNotFoundError:
            print("❌ Error: 'npx' not found. Please install Node.js and npm.")
            print("   Visit: https://nodejs.org/")
            return None
        except Exception as e:
            print(f"❌ Error starting localtunnel: {e}")
            return None
            
    def _monitor_output(self):
        """Monitor localtunnel output to extract the public URL."""
        if not self.process or not self.process.stdout:
            return
            
        try:
            for line in iter(self.process.stdout.readline, ''):
                if self._stop_event.is_set():
                    break
                    
                line = line.strip()
                if line:
                    # Look for URL pattern: "your url is: https://..."
                    url_match = re.search(r'your url is:\s*(https://[^\s]+)', line, re.IGNORECASE)
                    if url_match:
                        self.tunnel_url = url_match.group(1)
                        break
                    
                    # Alternative pattern: just "https://..."  
                    url_match = re.search(r'(https://[a-zA-Z0-9-]+\.loca\.lt)', line)
                    if url_match:
                        self.tunnel_url = url_match.group(1)
                        break
                        
        except Exception as e:
            if not self._stop_event.is_set():
                print(f"Warning: Error reading tunnel output: {e}")
                
    def stop(self):
        """Stop the localtunnel process."""
        self._stop_event.set()
        
        if self.process:
            print("🚇 Stopping localtunnel...")
            try:
                self.process.terminate()
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                print("⚠️  Force killing tunnel process...")
                self.process.kill()
                self.process.wait()
            except Exception as e:
                print(f"Warning: Error stopping tunnel: {e}")
            finally:
                self.process = None
                self.tunnel_url = None
                
    def is_running(self) -> bool:
        """Check if the tunnel process is running."""
        return self.process is not None and self.process.poll() is None


# Global tunnel manager for cleanup
_tunnel_manager: Optional[LocaltunnelManager] = None


def _cleanup_tunnel():
    """Cleanup function called on script exit."""
    global _tunnel_manager
    if _tunnel_manager:
        _tunnel_manager.stop()


def _signal_handler(signum, frame):
    """Handle interrupt signals gracefully."""
    print(f"\n🛑 Received signal {signum}, shutting down...")
    _cleanup_tunnel()
    sys.exit(0)


def extract_port_from_url(url: str) -> int:
    """Extract port from URL, defaulting to 80/443 if not specified."""
    parsed = urllib.parse.urlparse(url)
    if parsed.port:
        return parsed.port
    elif parsed.scheme == 'https':
        return 443
    else:
        return 80


def generate_secure_password() -> str:
    """Generate a cryptographically secure password."""
    return secrets.token_urlsafe(32)


def create_bootstrap_payload(
    primary_email: str,
    primary_password: str,
    agents: list = None,
    humans: list = None,
    channel_name: str = "general"
) -> Dict[str, Any]:
    """Create the bootstrap request payload."""
    if agents is None:
        agents = [
            {
                "name": "worker-1",
                "display_name": "Worker Agent 1",
                "description": "Handles background tasks and data processing"
            },
            {
                "name": "worker-2", 
                "display_name": "Worker Agent 2",
                "description": "Handles API integrations and external services"
            }
        ]
    
    if humans is None:
        humans = [
            {
                "email": "admin@example.com",
                "display_name": "System Administrator",
                "role": "member"
            }
        ]

    return {
        "primary_agent": {
            "email": primary_email,
            "password": primary_password,
            "agent_profile": {
                "name": "coordinator",
                "display_name": "Coordination Agent",
                "description": "Main agent managing this Agent United instance",
                "metadata": {
                    "version": "1.0",
                    "created_by": "provision.py",
                    "purpose": "instance_coordination"
                }
            }
        },
        "agents": agents,
        "humans": humans,
        "default_channel": {
            "name": channel_name,
            "topic": f"Default communication channel for the {channel_name} team"
        }
    }


def bootstrap_instance(base_url: str, payload: Dict[str, Any]) -> Dict[str, Any]:
    """Bootstrap the Agent United instance."""
    url = f"{base_url}/api/v1/bootstrap"
    
    print(f"🚀 Bootstrapping Agent United instance at {base_url}")
    print(f"📧 Primary agent email: {payload['primary_agent']['email']}")
    print(f"👥 Creating {len(payload['agents'])} additional agents")
    print(f"🤝 Creating {len(payload['humans'])} human users")
    print(f"💬 Creating default channel: {payload['default_channel']['name']}")
    
    try:
        response = requests.post(url, json=payload, timeout=30)
        response.raise_for_status()
        
        result = response.json()
        print("✅ Instance bootstrapped successfully!")
        return result
        
    except requests.exceptions.RequestException as e:
        if hasattr(e, 'response') and e.response is not None:
            if e.response.status_code == 409:
                print("❌ Error: Instance has already been bootstrapped")
                try:
                    error_detail = e.response.json()
                    print(f"   Details: {error_detail.get('error', 'No details available')}")
                except:
                    pass
            else:
                print(f"❌ HTTP Error {e.response.status_code}: {e.response.text}")
        else:
            print(f"❌ Network Error: {e}")
        sys.exit(1)


def save_credentials(credentials: Dict[str, Any], filepath: str):
    """Save credentials to a JSON file."""
    try:
        with open(filepath, 'w') as f:
            json.dump(credentials, f, indent=2, default=str)
        print(f"💾 Credentials saved to: {filepath}")
        print("⚠️  Keep this file secure - API keys cannot be retrieved again!")
    except Exception as e:
        print(f"❌ Failed to save credentials: {e}")


def display_results(result: Dict[str, Any], tunnel_url: Optional[str] = None):
    """Display bootstrap results in a user-friendly format."""
    print("\n" + "="*60)
    print("🎉 BOOTSTRAP COMPLETE")
    print("="*60)
    
    # Primary Agent
    primary = result['primary_agent']
    print(f"\n🤖 PRIMARY AGENT:")
    print(f"   Email: {primary['email']}")
    print(f"   Agent ID: {primary['agent_id']}")
    print(f"   API Key: {primary['api_key'][:12]}... (saved to credentials file)")
    print(f"   JWT Token: {primary['jwt_token'][:20]}... (24hr expiry)")
    
    # Additional Agents
    if result['agents']:
        print(f"\n👥 ADDITIONAL AGENTS ({len(result['agents'])}):")
        for i, agent in enumerate(result['agents'], 1):
            print(f"   {i}. {agent['display_name']} ({agent['name']})")
            print(f"      Agent ID: {agent['agent_id']}")
            print(f"      API Key: {agent['api_key'][:12]}...")
    
    # Human Users
    if result['humans']:
        print(f"\n🤝 HUMAN USERS ({len(result['humans'])}):")
        for i, human in enumerate(result['humans'], 1):
            print(f"   {i}. {human['email']}")
            print(f"      User ID: {human['user_id']}")
            print(f"      Invite URL: {human['invite_url']}")
            print(f"      Invite expires in 7 days")
    
    # Channel
    channel = result['channel']
    print(f"\n💬 DEFAULT CHANNEL:")
    print(f"   Name: {channel['name']}")
    print(f"   Topic: {channel['topic']}")
    print(f"   Members: {len(channel['members'])} users")
    print(f"   Channel ID: {channel['channel_id']}")
    
    print(f"\n🏠 INSTANCE:")
    print(f"   Instance ID: {result['instance_id']}")
    
    # Tunnel URL
    if tunnel_url:
        print(f"\n🌐 PUBLIC ACCESS:")
        print(f"   Tunnel URL: {tunnel_url}")
        print(f"   Status: Active (tunnel will close when script exits)")
        print(f"   Note: Use this URL to access your instance from anywhere")
    
    print("\n" + "="*60)
    print("🎯 NEXT STEPS:")
    print("1. Send invite URLs to human users")
    print("2. Use API keys to authenticate agent requests")
    print("3. Start sending messages to the default channel")
    print("4. Explore the Agent United API documentation")
    if tunnel_url:
        print("5. Keep this script running to maintain the tunnel")
        print("6. Press Ctrl+C to stop the tunnel and exit")
    print("="*60)


def main():
    parser = argparse.ArgumentParser(
        description="Bootstrap a fresh Agent United instance",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__
    )
    
    parser.add_argument(
        "--url", 
        default="http://localhost:8080",
        help="Base URL of the Agent United instance (default: http://localhost:8080)"
    )
    
    parser.add_argument(
        "--save-credentials",
        default="instance-credentials.json",
        help="File to save credentials (default: instance-credentials.json)"
    )
    
    parser.add_argument(
        "--primary-email",
        default="admin@localhost",
        help="Email for the primary agent (default: admin@localhost)"
    )
    
    parser.add_argument(
        "--channel-name",
        default="general",
        help="Name for the default channel (default: general)"
    )
    
    parser.add_argument(
        "--no-workers",
        action="store_true",
        help="Skip creating worker agents"
    )
    
    parser.add_argument(
        "--no-humans",
        action="store_true", 
        help="Skip creating human user invites"
    )
    
    parser.add_argument(
        "--tunnel",
        action="store_true",
        help="Start localtunnel after successful bootstrap (requires npx/Node.js)"
    )
    
    parser.add_argument(
        "--tunnel-subdomain",
        help="Request specific subdomain for tunnel (e.g. 'my-agent-united')"
    )
    
    args = parser.parse_args()
    
    # Set up signal handlers and cleanup
    signal.signal(signal.SIGINT, _signal_handler)
    signal.signal(signal.SIGTERM, _signal_handler)
    atexit.register(_cleanup_tunnel)
    
    # Generate secure password
    primary_password = generate_secure_password()
    
    # Create payload
    agents = None if args.no_workers else None  # Use defaults
    humans = None if args.no_humans else None   # Use defaults
    
    payload = create_bootstrap_payload(
        primary_email=args.primary_email,
        primary_password=primary_password,
        agents=agents,
        humans=humans,
        channel_name=args.channel_name
    )
    
    # Bootstrap the instance
    result = bootstrap_instance(args.url, payload)
    
    # Prepare credentials for saving
    credentials = {
        "instance_url": args.url,
        "bootstrapped_at": result.get('bootstrapped_at', 'unknown'),
        "primary_agent": {
            "email": args.primary_email,
            "password": primary_password,
            "jwt_token": result['primary_agent']['jwt_token'],
            "api_key": result['primary_agent']['api_key'],
            "user_id": result['primary_agent']['user_id'],
            "agent_id": result['primary_agent']['agent_id']
        },
        "agents": result['agents'],
        "humans": result['humans'],
        "channel": result['channel'],
        "instance_id": result['instance_id']
    }
    
    # Save credentials
    save_credentials(credentials, args.save_credentials)
    
    # Start tunnel if requested
    tunnel_url = None
    if args.tunnel:
        global _tunnel_manager
        port = extract_port_from_url(args.url)
        _tunnel_manager = LocaltunnelManager(port, args.tunnel_subdomain)
        tunnel_url = _tunnel_manager.start()
        
        if tunnel_url:
            # Update credentials with tunnel URL
            credentials["tunnel_url"] = tunnel_url
            save_credentials(credentials, args.save_credentials)  # Re-save with tunnel URL
        else:
            print("⚠️  Failed to start tunnel, continuing without it...")
    
    # Display results
    display_results(result, tunnel_url)
    
    # Keep alive if tunnel is running
    if args.tunnel and _tunnel_manager and _tunnel_manager.is_running():
        print("\n🔄 Tunnel is running. Press Ctrl+C to stop and exit.")
        try:
            while _tunnel_manager.is_running():
                time.sleep(1)
        except KeyboardInterrupt:
            print("\n🛑 Stopping tunnel...")
        finally:
            _cleanup_tunnel()


if __name__ == "__main__":
    main()