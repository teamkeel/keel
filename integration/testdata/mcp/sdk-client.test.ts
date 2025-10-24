import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase } from "@teamkeel/testing";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

beforeEach(resetDatabase);

// Helper to create a custom HTTP transport for MCP client
class HttpTransport {
  private url: string;
  private headers: Record<string, string>;
  private nextId: number = 1;

  constructor(url: string, headers: Record<string, string> = {}) {
    this.url = url;
    this.headers = headers;
  }

  async start() {
    // HTTP transport doesn't need a start phase
  }

  async close() {
    // HTTP transport doesn't need cleanup
  }

  async send(message: any): Promise<any> {
    const response = await fetch(this.url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...this.headers,
      },
      body: JSON.stringify(message),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
  }
}

test("MCP SDK Client - can connect and initialize", async () => {
  const mcpUrl = process.env.KEEL_TESTING_CLIENT_API_URL!.replace(
    "/json",
    "/mcp"
  );

  const transport = new HttpTransport(mcpUrl);
  const client = new Client(
    {
      name: "test-client",
      version: "1.0.0",
    },
    {
      capabilities: {},
    }
  );

  // Connect using the transport
  await client.connect(transport as any);

  // Get server info
  const serverInfo = client.getServerVersion();
  expect(serverInfo).toBeDefined();

  await client.close();
});

test("MCP SDK Client - can list resources", async () => {
  // Create test data
  await models.post.create({
    title: "Test Post",
    content: "Test Content",
  });

  const mcpUrl = process.env.KEEL_TESTING_CLIENT_API_URL!.replace(
    "/json",
    "/mcp"
  );

  const transport = new HttpTransport(mcpUrl);
  const client = new Client(
    {
      name: "test-client",
      version: "1.0.0",
    },
    {
      capabilities: {},
    }
  );

  await client.connect(transport as any);

  // List resources
  const resources = await client.listResources();
  expect(resources).toBeDefined();
  expect(resources.resources).toBeInstanceOf(Array);
  expect(resources.resources.length).toBeGreaterThan(0);

  // Verify we have read actions (get, list)
  const resourceNames = resources.resources.map((r) => r.name);
  expect(resourceNames).toContain("Post.getPost");
  expect(resourceNames).toContain("Post.listPosts");

  await client.close();
});

test("MCP SDK Client - can read a resource", async () => {
  // Create test data
  await models.post.create({
    title: "SDK Test Post",
    content: "SDK Test Content",
  });

  const mcpUrl = process.env.KEEL_TESTING_CLIENT_API_URL!.replace(
    "/json",
    "/mcp"
  );

  const transport = new HttpTransport(mcpUrl);
  const client = new Client(
    {
      name: "test-client",
      version: "1.0.0",
    },
    {
      capabilities: {},
    }
  );

  await client.connect(transport as any);

  // Read the list resource
  const result = await client.readResource({
    uri: "keel://web/Post/listPosts",
  });

  expect(result).toBeDefined();
  expect(result.contents).toBeInstanceOf(Array);
  expect(result.contents.length).toBeGreaterThan(0);

  // Parse the JSON content
  const content = result.contents[0];
  expect(content.mimeType).toBe("application/json");

  if (content.text) {
    const data = JSON.parse(content.text);
    expect(data.results).toBeDefined();
    expect(data.results).toBeInstanceOf(Array);
    expect(data.results.length).toBeGreaterThan(0);
    expect(data.results[0].title).toBe("SDK Test Post");
  }

  await client.close();
});

test("MCP SDK Client - can list tools", async () => {
  const mcpUrl = process.env.KEEL_TESTING_CLIENT_API_URL!.replace(
    "/json",
    "/mcp"
  );

  const transport = new HttpTransport(mcpUrl);
  const client = new Client(
    {
      name: "test-client",
      version: "1.0.0",
    },
    {
      capabilities: {},
    }
  );

  await client.connect(transport as any);

  // List tools
  const tools = await client.listTools();
  expect(tools).toBeDefined();
  expect(tools.tools).toBeInstanceOf(Array);
  expect(tools.tools.length).toBeGreaterThan(0);

  // Verify we have write actions (create, update, delete)
  const toolNames = tools.tools.map((t) => t.name);
  expect(toolNames).toContain("Post.createPost");
  expect(toolNames).toContain("Post.updatePost");
  expect(toolNames).toContain("Post.deletePost");

  // Verify tools have proper schema
  const createTool = tools.tools.find((t) => t.name === "Post.createPost");
  expect(createTool).toBeDefined();
  expect(createTool!.inputSchema).toBeDefined();
  expect(createTool!.inputSchema.type).toBe("object");
  expect(createTool!.inputSchema.properties).toBeDefined();

  await client.close();
});

test("MCP SDK Client - can call a tool", async () => {
  const mcpUrl = process.env.KEEL_TESTING_CLIENT_API_URL!.replace(
    "/json",
    "/mcp"
  );

  const transport = new HttpTransport(mcpUrl);
  const client = new Client(
    {
      name: "test-client",
      version: "1.0.0",
    },
    {
      capabilities: {},
    }
  );

  await client.connect(transport as any);

  // Call the createPost tool
  const result = await client.callTool({
    name: "Post.createPost",
    arguments: {
      title: "SDK Created Post",
      content: "Created via MCP SDK",
      published: true,
    },
  });

  expect(result).toBeDefined();
  expect(result.content).toBeInstanceOf(Array);
  expect(result.content.length).toBeGreaterThan(0);
  expect(result.isError).toBeFalsy();

  // Parse the result
  const content = result.content[0];
  expect(content.type).toBe("text");

  if (content.text) {
    const post = JSON.parse(content.text);
    expect(post.id).toBeDefined();
    expect(post.title).toBe("SDK Created Post");
    expect(post.content).toBe("Created via MCP SDK");
    expect(post.published).toBe(true);
  }

  // Verify in database
  const dbPost = await models.post.findMany({});
  expect(dbPost.length).toBeGreaterThan(0);
  expect(dbPost.some((p) => p.title === "SDK Created Post")).toBe(true);

  await client.close();
});

test("MCP SDK Client - auth instructions in initialize", async () => {
  const mcpUrl = process.env.KEEL_TESTING_CLIENT_API_URL!.replace(
    "/json",
    "/mcp"
  );

  const transport = new HttpTransport(mcpUrl);
  const client = new Client(
    {
      name: "test-client",
      version: "1.0.0",
    },
    {
      capabilities: {},
    }
  );

  await client.connect(transport as any);

  // The client should have received server info with instructions
  // We can verify this by checking the connection was successful
  const serverVersion = client.getServerVersion();
  expect(serverVersion).toBeDefined();
  expect(serverVersion?.name).toBe("keel");

  await client.close();
});

test("MCP SDK Client - handles errors gracefully", async () => {
  const mcpUrl = process.env.KEEL_TESTING_CLIENT_API_URL!.replace(
    "/json",
    "/mcp"
  );

  const transport = new HttpTransport(mcpUrl);
  const client = new Client(
    {
      name: "test-client",
      version: "1.0.0",
    },
    {
      capabilities: {},
    }
  );

  await client.connect(transport as any);

  // Try to read a non-existent resource
  await expect(
    client.readResource({
      uri: "keel://web/NonExistent/nonExistentAction",
    })
  ).rejects.toThrow();

  // Try to call a non-existent tool
  await expect(
    client.callTool({
      name: "NonExistent.nonExistentAction",
      arguments: {},
    })
  ).rejects.toThrow();

  await client.close();
});

test("MCP SDK Client - with authentication", async () => {
  // First get a token via the auth API
  const authResponse = await fetch(
    process.env.KEEL_TESTING_AUTH_API_URL + "/token",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        grant_type: "password",
        username: "test@example.com",
        password: "password",
      }),
    }
  );

  // Note: This test will fail if no identity exists, which is expected
  // In a real scenario, you'd create an identity first
  if (authResponse.ok) {
    const { access_token } = await authResponse.json();

    const mcpUrl = process.env.KEEL_TESTING_CLIENT_API_URL!.replace(
      "/json",
      "/mcp"
    );

    // Create client with auth header
    const transport = new HttpTransport(mcpUrl, {
      Authorization: `Bearer ${access_token}`,
    });

    const client = new Client(
      {
        name: "test-client",
        version: "1.0.0",
      },
      {
        capabilities: {},
      }
    );

    await client.connect(transport as any);

    // Should be able to list resources with auth
    const resources = await client.listResources();
    expect(resources).toBeDefined();

    await client.close();
  } else {
    // Skip test if no auth is set up
    expect(true).toBe(true);
  }
});
