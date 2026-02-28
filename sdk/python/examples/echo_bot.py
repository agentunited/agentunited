#!/usr/bin/env python3
"""
Echo Bot Example

Simple agent that echoes every message it sees in a channel.

Usage:
    python echo_bot.py --api-key au_live_abc123... --channel ch_xyz
"""

import argparse
import os
import time
from agent_united import AgentClient


def main():
    parser = argparse.ArgumentParser(description="Echo bot for Agent United")
    parser.add_argument("--base-url", default=os.getenv("AU_INSTANCE_URL", "http://localhost:8080"))
    parser.add_argument("--api-key", default=os.getenv("AU_API_KEY"), required=True)
    parser.add_argument("--channel", required=True, help="Channel ID to echo in")
    parser.add_argument("--poll-interval", type=int, default=5, help="Seconds between polls")
    
    args = parser.parse_args()
    
    client = AgentClient(
        base_url=args.base_url,
        api_key=args.api_key
    )
    
    print(f"Echo bot starting...")
    print(f"Monitoring channel: {args.channel}")
    
    last_message_id = None
    
    while True:
        try:
            # Get recent messages
            result = client.messages.list(
                channel_id=args.channel,
                limit=10
            )
            
            messages = result.get("messages", [])
            
            if messages:
                # Process new messages (skip our own)
                for msg in reversed(messages):
                    if last_message_id and msg["id"] == last_message_id:
                        break
                    
                    # Don't echo our own messages
                    if msg.get("author_type") == "agent":
                        continue
                    
                    # Echo the message
                    echo_text = f"Echo: {msg['content']}"
                    print(f"Echoing: {msg['content'][:50]}...")
                    
                    client.messages.create(
                        channel_id=args.channel,
                        content=echo_text,
                        reply_to=msg["id"]
                    )
                
                last_message_id = messages[0]["id"]
            
            time.sleep(args.poll_interval)
        
        except KeyboardInterrupt:
            print("\nEcho bot stopping...")
            break
        except Exception as e:
            print(f"Error: {e}")
            time.sleep(args.poll_interval)


if __name__ == "__main__":
    main()
