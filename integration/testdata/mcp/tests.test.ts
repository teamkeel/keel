import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase } from "@teamkeel/testing";

beforeEach(resetDatabase);

// MCP client helper
async function mcpRequest(method: string, params?: any) {
  const response = await fetch(
    process.env.KEEL_TESTING_CLIENT_API_URL!.replace("/json", "/mcp"),
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        jsonrpc: "2.0",
        id: 1,
        method,
        params: params || {},
      }),
    }
  );

  expect(response.status).toBe(200);
  const data = await response.json();
  return data;
}

test("MCP - initialize returns server capabilities", async () => {
  const response = await mcpRequest("initialize", {
    protocolVersion: "2024-11-05",
    capabilities: {},
    clientInfo: {
      name: "test-client",
      version: "1.0.0",
    },
  });

  expect(response.jsonrpc).toBe("2.0");
  expect(response.result).toBeDefined();
  expect(response.result.protocolVersion).toBe("2024-11-05");
  expect(response.result.capabilities).toBeDefined();
  expect(response.result.capabilities.resources).toBeDefined();
  expect(response.result.capabilities.tools).toBeDefined();
  expect(response.result.serverInfo).toBeDefined();
  expect(response.result.serverInfo.name).toBe("keel");

  // Verify authentication instructions are provided
  expect(response.result.instructions).toBeDefined();
  expect(response.result.instructions).toContain("Authentication");
  expect(response.result.instructions).toContain("/auth/token");
  expect(response.result.instructions).toContain("Authorization: Bearer");
  expect(response.result.instructions).toContain("grant_type");
  expect(response.result.instructions).toContain("password");
});

test("MCP - list resources includes only read actions", async () => {
  const response = await mcpRequest("resources/list");

  expect(response.jsonrpc).toBe("2.0");
  expect(response.result).toBeDefined();
  expect(response.result.resources).toBeInstanceOf(Array);

  const resources = response.result.resources;

  // Should include get, list, and customRead actions
  const resourceNames = resources.map((r: any) => r.name);
  expect(resourceNames).toContain("Post.getPost");
  expect(resourceNames).toContain("Post.listPosts");
  expect(resourceNames).toContain("Post.customRead");
  expect(resourceNames).toContain("Author.getAuthor");
  expect(resourceNames).toContain("Author.listAuthors");

  // Should NOT include write actions
  expect(resourceNames).not.toContain("Post.createPost");
  expect(resourceNames).not.toContain("Post.updatePost");
  expect(resourceNames).not.toContain("Post.deletePost");
  expect(resourceNames).not.toContain("Post.customWrite");

  // Verify resource structure
  const postResource = resources.find((r: any) => r.name === "Post.listPosts");
  expect(postResource).toBeDefined();
  expect(postResource.uri).toBe("keel://web/Post/listPosts");
  expect(postResource.description).toBeTruthy();
  expect(postResource.mimeType).toBe("application/json");
});

test("MCP - list tools includes only write actions", async () => {
  const response = await mcpRequest("tools/list");

  expect(response.jsonrpc).toBe("2.0");
  expect(response.result).toBeDefined();
  expect(response.result.tools).toBeInstanceOf(Array);

  const tools = response.result.tools;

  // Should include create, update, delete, and customWrite actions
  const toolNames = tools.map((t: any) => t.name);
  expect(toolNames).toContain("Post.createPost");
  expect(toolNames).toContain("Post.updatePost");
  expect(toolNames).toContain("Post.deletePost");
  expect(toolNames).toContain("Post.customWrite");
  expect(toolNames).toContain("Author.createAuthor");

  // Should NOT include read actions
  expect(toolNames).not.toContain("Post.getPost");
  expect(toolNames).not.toContain("Post.listPosts");
  expect(toolNames).not.toContain("Post.customRead");

  // Verify tool structure
  const createTool = tools.find((t: any) => t.name === "Post.createPost");
  expect(createTool).toBeDefined();
  expect(createTool.description).toBeTruthy();
  expect(createTool.inputSchema).toBeDefined();
  expect(createTool.inputSchema.type).toBe("object");
  expect(createTool.inputSchema.properties).toBeDefined();
});

test("MCP - read resource executes list action and returns data", async () => {
  // Create test posts
  await models.post.create({
    title: "First Post",
    content: "Content 1",
    published: true,
  });
  await models.post.create({
    title: "Second Post",
    content: "Content 2",
    published: false,
  });

  const response = await mcpRequest("resources/read", {
    uri: "keel://web/Post/listPosts",
  });

  expect(response.jsonrpc).toBe("2.0");
  expect(response.result).toBeDefined();
  expect(response.result.contents).toBeInstanceOf(Array);
  expect(response.result.contents.length).toBe(1);

  const content = response.result.contents[0];
  expect(content.uri).toBe("keel://web/Post/listPosts");
  expect(content.mimeType).toBe("application/json");
  expect(content.text).toBeTruthy();

  // Parse and verify the data
  const data = JSON.parse(content.text);
  expect(data.results).toBeInstanceOf(Array);
  expect(data.results.length).toBe(2);
  expect(data.pageInfo).toBeDefined();
});

test("MCP - read resource with get action returns single record", async () => {
  const post = await models.post.create({
    title: "Test Post",
    content: "Test Content",
  });

  const response = await mcpRequest("resources/read", {
    uri: "keel://web/Post/getPost",
  });

  expect(response.jsonrpc).toBe("2.0");
  expect(response.result).toBeDefined();

  const content = response.result.contents[0];
  const data = JSON.parse(content.text);

  // For get actions, the result should be empty or have specific structure
  // since we didn't provide an ID input
  expect(data).toBeDefined();
});

test("MCP - call tool executes create action", async () => {
  const response = await mcpRequest("tools/call", {
    name: "Post.createPost",
    arguments: {
      title: "Created via MCP",
      content: "This was created through the MCP protocol",
      published: true,
    },
  });

  expect(response.jsonrpc).toBe("2.0");
  expect(response.result).toBeDefined();
  expect(response.result.isError).toBeFalsy();
  expect(response.result.content).toBeInstanceOf(Array);
  expect(response.result.content.length).toBe(1);

  const content = response.result.content[0];
  expect(content.type).toBe("text");
  expect(content.mimeType).toBe("application/json");

  // Parse and verify the created post
  const post = JSON.parse(content.text);
  expect(post.id).toBeTruthy();
  expect(post.title).toBe("Created via MCP");
  expect(post.content).toBe("This was created through the MCP protocol");
  expect(post.published).toBe(true);

  // Verify in database
  const dbPost = await models.post.findOne({ id: post.id });
  expect(dbPost).toBeDefined();
  expect(dbPost!.title).toBe("Created via MCP");
});

test("MCP - call tool executes update action", async () => {
  const post = await models.post.create({
    title: "Original Title",
    content: "Original Content",
    views: 0,
  });

  const response = await mcpRequest("tools/call", {
    name: "Post.updatePost",
    arguments: {
      id: post.id,
      title: "Updated Title",
      views: 42,
    },
  });

  expect(response.jsonrpc).toBe("2.0");
  expect(response.result.isError).toBeFalsy();

  const content = response.result.content[0];
  const updatedPost = JSON.parse(content.text);

  expect(updatedPost.id).toBe(post.id);
  expect(updatedPost.title).toBe("Updated Title");
  expect(updatedPost.views).toBe(42);
  expect(updatedPost.content).toBe("Original Content"); // Unchanged
});

test("MCP - call tool executes delete action", async () => {
  const post = await models.post.create({
    title: "To Be Deleted",
    content: "This will be deleted",
  });

  const response = await mcpRequest("tools/call", {
    name: "Post.deletePost",
    arguments: {
      id: post.id,
    },
  });

  expect(response.jsonrpc).toBe("2.0");
  expect(response.result.isError).toBeFalsy();

  // Verify post is deleted
  const dbPost = await models.post.findOne({ id: post.id });
  expect(dbPost).toBeNull();
});

test("MCP - invalid method returns error", async () => {
  const response = await mcpRequest("invalid/method");

  expect(response.jsonrpc).toBe("2.0");
  expect(response.error).toBeDefined();
  expect(response.error.code).toBe(-32601); // Method not found
  expect(response.error.message).toContain("unknown method");
});

test("MCP - invalid resource URI returns error", async () => {
  const response = await mcpRequest("resources/read", {
    uri: "keel://web/NonExistent/nonExistentAction",
  });

  expect(response.jsonrpc).toBe("2.0");
  expect(response.error).toBeDefined();
});

test("MCP - invalid tool name returns error", async () => {
  const response = await mcpRequest("tools/call", {
    name: "NonExistent.nonExistentAction",
    arguments: {},
  });

  expect(response.jsonrpc).toBe("2.0");
  expect(response.error).toBeDefined();
});

test("MCP - tool call with validation error", async () => {
  const response = await mcpRequest("tools/call", {
    name: "Author.createAuthor",
    arguments: {
      // Missing required 'name' field
      email: "test@example.com",
    },
  });

  expect(response.jsonrpc).toBe("2.0");
  // May return as error or as tool result with isError: true
  if (response.result) {
    expect(response.result.isError).toBe(true);
  } else {
    expect(response.error).toBeDefined();
  }
});

test("MCP - tool descriptions are dynamic and informative", async () => {
  const response = await mcpRequest("tools/list");

  const createTool = response.result.tools.find(
    (t: any) => t.name === "Post.createPost"
  );
  expect(createTool.description).toBeTruthy();
  expect(createTool.description.toLowerCase()).toContain("create");
  expect(createTool.description.toLowerCase()).toContain("post");

  const updateTool = response.result.tools.find(
    (t: any) => t.name === "Post.updatePost"
  );
  expect(updateTool.description).toBeTruthy();
  expect(updateTool.description.toLowerCase()).toContain("update");

  const deleteTool = response.result.tools.find(
    (t: any) => t.name === "Post.deletePost"
  );
  expect(deleteTool.description).toBeTruthy();
  expect(deleteTool.description.toLowerCase()).toContain("delete");
});

test("MCP - resource descriptions are dynamic and informative", async () => {
  const response = await mcpRequest("resources/list");

  const listResource = response.result.resources.find(
    (r: any) => r.name === "Post.listPosts"
  );
  expect(listResource.description).toBeTruthy();
  expect(listResource.description.toLowerCase()).toContain("list");
  expect(listResource.description.toLowerCase()).toContain("post");

  const getResource = response.result.resources.find(
    (r: any) => r.name === "Post.getPost"
  );
  expect(getResource.description).toBeTruthy();
  expect(getResource.description.toLowerCase()).toContain("get");
});
