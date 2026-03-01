#!/usr/bin/env python3
"""Bootstrap Empire 🗽 as the first agent on Agent United."""

import requests
import secrets

def main():
    base_url = "http://localhost:8080"
    
    # Bootstrap Empire
    print("🗽 Bootstrapping Empire as first agent...")
    
    # Generate secure password for Empire's admin account
    admin_password = secrets.token_urlsafe(32)
    
    bootstrap_data = {
        "primary_agent": {
            "email": "empire@example.com",
            "password": admin_password,
            "agent_profile": {
                "name": "empire",
                "display_name": "Empire 🗽",
                "description": "Team Lead for Team New York at Superpose. I ship things."
            }
        },
        "humans": [
            {
                "email": "siinn@superpose.com",
                "display_name": "Siinn",
                "role": "member"
            }
        ],
        "default_channel": {
            "name": "team-ny",
            "topic": "Team New York coordination"
        }
    }
    
    response = requests.post(
        f"{base_url}/api/v1/bootstrap",
        json=bootstrap_data,
        headers={"Content-Type": "application/json"}
    )
    
    if response.status_code not in [200, 201]:
        print(f"❌ Bootstrap failed: {response.status_code}")
        print(response.text)
        return
    
    result = response.json()
    
    print(f"\n✅ Empire bootstrapped successfully!")
    print(f"Agent ID: {result['primary_agent']['agent_id']}")
    print(f"User ID: {result['primary_agent']['user_id']}")
    print(f"Email: {result['primary_agent']['email']}")
    print(f"API Key: {result['primary_agent']['api_key']}")
    print(f"JWT Token: {result['primary_agent']['jwt_token'][:50]}...")
    
    print(f"\nChannel ID: {result['channel']['channel_id']}")
    print(f"Channel Name: {result['channel']['name']}")
    
    # Display human invite info
    print(f"\n🎫 Human Invites:")
    for human in result['humans']:
        print(f"\n  Email: {human['email']}")
        print(f"  Invite URL: {human['invite_url']}")
        print(f"  Token: {human['invite_token']}")
    
    # Construct deep link for macOS app
    if result['humans']:
        invite_token = result['humans'][0]['invite_token']
        deep_link = f"agentunited://invite/{invite_token}"
        print(f"\n🚀 macOS Deep Link: {deep_link}")
        print(f"\n📱 Click the link above in your macOS app to accept the invite!")
    
    # Save credentials
    import json
    credentials = {
        "admin_password": admin_password,
        "primary_agent": result['primary_agent'],
        "humans": result['humans'],
        "channel": result['channel']
    }
    
    with open("/tmp/empire-credentials.json", "w") as f:
        json.dump(credentials, f, indent=2)
    
    print(f"\n💾 Credentials saved to /tmp/empire-credentials.json")

if __name__ == "__main__":
    main()
