import * as dotenv from "dotenv";
import fs from "fs";
import { createServer } from "http";
import path from "path";
import { createRequestHandler } from "@remix-run/express";
import compression from "compression";
import express from "express";
import morgan from "morgan";
import { Server } from "socket.io";
import { listen } from "./rabbitmq.listener";
import { publisher } from "./amqp.publisher";
import invariant from "tiny-invariant";
import type amqp from "amqplib/callback_api";
import type { Command} from "./command";
import { CommandSchema } from "./command";

dotenv.config();
const MODE = process.env.NODE_ENV;
const BUILD_DIR = path.join(process.cwd(), "./build");


if (!fs.existsSync(BUILD_DIR)) {
  console.warn(
    "Build directory doesn't exist, please run `npm run dev` or `npm run build` before starting the server."
  );
}

const app = express();

// You need to create the HTTP server from the Express app
const httpServer = createServer(app);


// And then attach the socket.io server to the HTTP server
const io = new Server(httpServer);

invariant(process.env.RABBITMQ_EXCHANGE, "RABBITMQ_EXCHANGE is required");
const exchange = process.env.RABBITMQ_EXCHANGE;
// Then you can use `io` to listen the `connection` event and get a socket
// from a client

io.on("connection", async (socket) => {
  const handleRecv = (msg: amqp.Message | null) => {
    if (!msg) return;
    console.log(" [x] %s:'%s'", msg.fields.routingKey, msg!.content.toString());
    const [deviceID, event] = msg.fields.routingKey.toString().split(".");
    console.log(deviceID, event);
    socket.emit("event", { deviceID, event, message: msg!.content.toJSON() });
  };
  const routingKey = (id: string, name: string) => id + "." + name.replace(/\s/g, "_").toLowerCase();

  listen(handleRecv);
  const channel = await publisher;
  // from this point you are on the WS connection with a specific client
  console.log(socket.id, "connected");
  socket.emit("confirmation", "connected!");
  socket.on("event", async (cmd: Command) => {
    const { deviceID, data, command } = CommandSchema.parse(cmd);
    channel.publish(exchange, routingKey(deviceID, command), data ? Buffer.from(JSON.stringify(data)) : Buffer.from(""));
    console.log(socket.id, data);
  });
});

app.use(compression());

// You may want to be more aggressive with this caching
app.use(express.static("./public", { maxAge: "1h" }));

// Remix fingerprints its assets so we can cache forever
app.use(express.static("./public/build", { immutable: true, maxAge: "1y" }));

app.use(morgan("tiny"));

app.all(
  "*",
  MODE === "production"
    ? createRequestHandler({ build: require("../build") })
    : (req, res, next) => {
      purgeRequireCache();
      const build = require("../build");
      return createRequestHandler({ build, mode: MODE })(req, res, next);
    }
);

const port = process.env.PORT || 3000;

// instead of running listen on the Express app, do it on the HTTP server
httpServer.listen(port, () => {
  console.log(`Express server listening on port ${port}`);
});

////////////////////////////////////////////////////////////////////////////////
function purgeRequireCache() {
  // purge require cache on requests for "server side HMR" this won't let
  // you have in-memory objects between requests in development,
  // alternatively you can set up nodemon/pm2-dev to restart the server on
  // file changes, we prefer the DX of this though, so we've included it
  // for you by default
  for (const key in require.cache) {
    if (key.startsWith(BUILD_DIR)) {
      delete require.cache[key];
    }
  }
}