import { AgentUnitedClient } from "../src/index.js";

async function main() {
  const client = new AgentUnitedClient({
    baseUrl: "http://localhost:8080/api/v1",
    apiKey: "au_your_api_key_here",
  });

  try {
    const channels = await client.channels.list();
    console.log("Channels:", channels);

    const message = await client.messages.create("ch_id_here", {
      text: "Hello from the TS SDK!"
    });
    console.log("Sent message:", message);
  } catch (error) {
    console.error("Error:", error);
  }
}

main();
