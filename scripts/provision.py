#!/usr/bin/env python3
"""
Agent United Provisioning Script

This script provisions a fresh Agent United instance via the Bootstrap API.
Designed to be run by AI agents for autonomous setup.

Usage:
    python provision.py --config config.json
    python provision.py --interactive
"""

import argparse
import json
import os
import secrets
import subprocess
import sys
import time
from pathlib import Path
from typing import Optional

import requests


class Colors:
    """ANSI color codes for terminal output"""
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'


def print_header(msg: str):
    print(f"\n{Colors.HEADER}{Colors.BOLD}{msg}{Colors.ENDC}")


def print_success(msg: str):
    print(f"{Colors.OKGREEN}✓ {msg}{Colors.ENDC}")


def print_error(msg: str):
    print(f"{Colors.FAIL}✗ {msg}{Colors.ENDC}", file=sys.stderr)


def print_info(msg: str):
    print(f"{Colors.OKCYAN}ℹ {msg}{Colors.ENDC}")


def generate_secure_password(length: int = 32) -> str:
    """Generate a cryptographically secure password"""
    return secrets.token_urlsafe(length)


def wait_for_api_health(base_url: str, max_retries: int = 30, retry_delay: int = 2) -> bool:
    """Wait for API health check to pass"""
    print_info(f"Waiting for API at {base_url}/health...")
    
    for i in range(max_retries):
        try:
            response = requests.get(f"{base_url}/health", timeout=5)
            if response.status_code == 200:
                print_success("API is healthy")
                return True
        except requests.RequestException:
            pass
        
        if i < max_retries - 1:
            time.sleep(retry_delay)
    
    return False


def bootstrap_instance(base_url: str, config: dict) -> dict:
    """Call the bootstrap API to provision the instance"""
    print_header("Provisioning Instance")
    
    endpoint = f"{base_url}/api/v1/bootstrap"
    
    print_info(f"Calling {endpoint}...")
    
    try:
        response = requests.post(
            endpoint,
            json=config,
            headers={"Content-Type": "application/json"},
            timeout=30
        )
        
        if response.status_code == 201:
            print_success("Instance provisioned successfully")
            return response.json()
        elif response.status_code == 409:
            print_error("Instance already bootstrapped (users exist)")
            sys.exit(1)
        else:
            print_error(f"Bootstrap failed: {response.status_code}")
            print_error(response.text)
            sys.exit(1)
    
    except requests.RequestException as e:
        print_error(f"Failed to connect to API: {e}")
        sys.exit(1)


def save_credentials(result: dict, output_path: Path):
    """Save bootstrap result to file"""
    print_header("Saving Credentials")
    
    with open(output_path, 'w') as f:
        json.dump(result, f, indent=2)
    
    # Set restrictive permissions (owner read/write only)
    os.chmod(output_path, 0o600)
    
    print_success(f"Credentials saved to: {output_path}")
    print_info("⚠️  Keep this file secure. API keys cannot be retrieved again.")


def download_macos_app(download_url: str, install_path: Path) -> bool:
    """Download and install macOS app"""
    print_header("Installing macOS App")
    
    try:
        # Download .dmg
        print_info(f"Downloading from {download_url}...")
        dmg_path = Path("/tmp/AgentUnited.dmg")
        
        response = requests.get(download_url, stream=True, timeout=60)
        response.raise_for_status()
        
        with open(dmg_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
        
        print_success("Download complete")
        
        # Mount DMG
        print_info("Mounting disk image...")
        subprocess.run(["hdiutil", "attach", str(dmg_path), "-nobrowse"], check=True, capture_output=True)
        
        # Copy app to Applications
        print_info("Installing to /Applications...")
        app_src = Path("/Volumes/Agent United/Agent United.app")
        
        if app_src.exists():
            subprocess.run(["cp", "-R", str(app_src), str(install_path)], check=True)
            print_success(f"Installed to {install_path}")
        else:
            print_error(f"App not found at {app_src}")
            return False
        
        # Unmount DMG
        print_info("Cleaning up...")
        subprocess.run(["hdiutil", "detach", "/Volumes/Agent United"], check=True, capture_output=True)
        dmg_path.unlink()
        
        return True
    
    except Exception as e:
        print_error(f"Failed to install macOS app: {e}")
        return False


def open_macos_app(app_path: Path, jwt_token: str):
    """Open macOS app with auto-login deep link"""
    print_info("Opening Agent United app...")
    
    # Deep link with JWT for auto-login
    deep_link = f"agentunited://auto-login?token={jwt_token}"
    
    try:
        subprocess.run(["open", "-a", str(app_path), "--args", deep_link], check=True)
        print_success("App opened")
    except Exception as e:
        print_error(f"Failed to open app: {e}")
        print_info(f"You can manually open: {app_path}")


def print_summary(result: dict, credentials_path: Path):
    """Print provisioning summary"""
    print_header("Provisioning Complete")
    
    primary_agent = result.get("primary_agent", {})
    agents = result.get("agents", [])
    humans = result.get("humans", [])
    
    print(f"\n{Colors.BOLD}Primary Agent:{Colors.ENDC}")
    print(f"  Email: {primary_agent.get('email')}")
    print(f"  API Key: {primary_agent.get('api_key')[:20]}... (full key in {credentials_path})")
    
    if agents:
        print(f"\n{Colors.BOLD}Additional Agents:{Colors.ENDC}")
        for agent in agents:
            print(f"  - {agent.get('name')}: {agent.get('api_key')[:20]}...")
    
    if humans:
        print(f"\n{Colors.BOLD}Human Invites:{Colors.ENDC}")
        for human in humans:
            print(f"  - {human.get('email')}")
            print(f"    Invite: {human.get('invite_url')}")
    
    print(f"\n{Colors.BOLD}Next Steps:{Colors.ENDC}")
    print(f"  1. Store credentials securely: {credentials_path}")
    print(f"  2. Send invite URLs to humans")
    print(f"  3. Start using API with your API key")


def main():
    parser = argparse.ArgumentParser(description="Provision Agent United instance")
    parser.add_argument("--config", type=Path, help="Path to bootstrap config JSON")
    parser.add_argument("--interactive", action="store_true", help="Interactive mode (prompts for input)")
    parser.add_argument("--base-url", default="http://localhost:8080", help="API base URL")
    parser.add_argument("--output", type=Path, default=Path("instance-credentials.json"), help="Output credentials file")
    parser.add_argument("--install-macos", choices=["auto", "manual", "skip"], default="auto", help="macOS app installation method")
    parser.add_argument("--macos-download-url", default="https://agentunited.ai/download/macos", help="macOS app download URL")
    
    args = parser.parse_args()
    
    print_header("Agent United Provisioning Script")
    
    # Load or create config
    if args.config:
        print_info(f"Loading config from {args.config}")
        with open(args.config) as f:
            config = json.load(f)
    elif args.interactive:
        print_error("Interactive mode not yet implemented")
        print_info("Use --config with a JSON file for now")
        sys.exit(1)
    else:
        # Default config for demo
        config = {
            "primary_agent": {
                "email": "admin@agentunited.local",
                "password": generate_secure_password(),
                "agent_profile": {
                    "name": "coordinator",
                    "display_name": "Coordination Agent",
                    "description": "Main agent managing this instance"
                }
            },
            "agents": [],
            "humans": [],
            "default_channel": {
                "name": "general",
                "topic": "Agent coordination channel"
            }
        }
        print_info("Using default config (coordinator agent, no additional agents/humans)")
    
    # Wait for API to be ready
    if not wait_for_api_health(args.base_url):
        print_error("API health check failed")
        sys.exit(1)
    
    # Bootstrap instance
    result = bootstrap_instance(args.base_url, config)
    
    # Save credentials
    save_credentials(result, args.output)
    
    # Install macOS app
    if sys.platform == "darwin" and args.install_macos != "skip":
        if args.install_macos == "auto":
            app_path = Path("/Applications/Agent United.app")
            if download_macos_app(args.macos_download_url, app_path):
                # Open app with auto-login
                jwt_token = result["primary_agent"]["jwt_token"]
                open_macos_app(app_path, jwt_token)
        elif args.install_macos == "manual":
            print_info(f"Download macOS app: {args.macos_download_url}")
    
    # Print summary
    print_summary(result, args.output)


if __name__ == "__main__":
    main()
