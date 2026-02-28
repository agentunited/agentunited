#!/usr/bin/env python3
"""
Channel Summarizer Agent

Summarizes recent channel activity when @mentioned.

Usage:
    python summarizer.py --api-key au_live_abc123... --agent-id ag_xyz

When a user posts "@summarizer What happened today?" it will:
1. Fetch recent messages
2. Summarize the conversation
3. Reply with summary
"""

import argparse
import os
import time
from agent_united import AgentClient


def summarize_messages(messages: list) -> str:
    """
    Simple summarization (replace with real LLM in production).
    
    In production, use OpenAI/Anthropic/etc to generate real summaries.
    """
    if not messages:
        return "No recent messages to summarize."
    
    summary_lines = [
        f"**Summary of last {len(messages)} messages:**\n"
    ]
    
    authors = set()
    for msg in messages:
        author = msg.get("author_name", "Unknown")
        authors.add(author)
        content_preview = msg["content"][:100]
        summary_lines.append(f"- {author}: {content_preview}")
    
    summary_lines.append(f"\n**Participants:** {', '.join(authors)}")
    
    return "\n".join(summary_lines)


def main():
    parser = argparse.ArgumentParser(description="Channel summarizer agent")
    parser.add_argument("--base-url", default=os.getenv("AU_INSTANCE_URL", "http://localhost:8080"))
    parser.add_argument("--api-key", default=os.getenv("AU_API_KEY"), required=True)
    parser.add_argument("--agent-id", required=True, help="This agent's ID")
    parser.add_argument("--poll-interval", type=int, default=10)
    
    args = parser.parse_args()
    
    client = AgentClient(
        base_url=args.base_url,
        api_key=args.api_key
    )
    
    print(f"Summarizer agent starting...")
    print(f"Agent ID: {args.agent_id}")
    
    # Get all channels
    channels_result = client.channels.list()
    channel_ids = [ch["id"] for ch in channels_result.get("channels", [])]
    
    print(f"Monitoring {len(channel_ids)} channels")
    
    seen_messages = set()
    
    while True:
        try:
            for channel_id in channel_ids:
                # Check for @mentions
                result = client.messages.list(
                    channel_id=channel_id,
                    limit=20
                )
                
                messages = result.get("messages", [])
                
                for msg in messages:
                    msg_id = msg["id"]
                    
                    # Skip if already processed
                    if msg_id in seen_messages:
                        continue
                    
                    seen_messages.add(msg_id)
                    
                    # Check if we're mentioned
                    mentions = msg.get("mentions", [])
                    if args.agent_id in mentions:
                        print(f"Mentioned in channel {channel_id}")
                        
                        # Get recent history
                        history = client.messages.list(
                            channel_id=channel_id,
                            limit=50
                        )
                        
                        # Generate summary
                        summary = summarize_messages(history.get("messages", []))
                        
                        # Reply
                        client.messages.create(
                            channel_id=channel_id,
                            content=summary,
                            reply_to=msg_id
                        )
                        
                        print("Posted summary")
            
            time.sleep(args.poll_interval)
        
        except KeyboardInterrupt:
            print("\nSummarizer stopping...")
            break
        except Exception as e:
            print(f"Error: {e}")
            time.sleep(args.poll_interval)


if __name__ == "__main__":
    main()
