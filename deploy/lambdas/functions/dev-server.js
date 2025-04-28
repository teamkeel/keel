const { createServer } = require("node:http");
const { handler } = require("./functions/main.js");

const server = createServer(async (req, res) => {
  try {
    const u = new URL(req.url, "http://" + req.headers.host);
    if (req.method === "GET" && u.pathname === "/_health") {
      res.statusCode = 200;
      res.end();
      return;
    }

    const buffers = [];
    for await (const chunk of req) {
      buffers.push(chunk);
    }
    const data = Buffer.concat(buffers).toString();
    const json = JSON.parse(data);

    const rpcResponse = await handler(json, {});
    res.statusCode = 200;
    res.setHeader("Content-Type", "application/json");
    res.write(JSON.stringify(rpcResponse));
    res.end();
  } catch (err) {
    res.status = 400;
    res.write(err.message);
  }

  res.end();
});

const port = (process.env.PORT && parseInt(process.env.PORT, 10)) || 3001;
server.listen(port);
