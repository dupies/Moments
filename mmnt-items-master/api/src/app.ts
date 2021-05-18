import "express-async-errors";

import path from "path";

import express, { Request, Response } from "express";

import cors from "cors";

import { json, urlencoded } from "body-parser";

import initMomentsRouter from "./routes/moment";
import initMmntItemsRouter from "./routes/mmnt-items";
import initMarketRouter from "./routes/market";
import { MomentsService } from "./services/moment";
import { MmntItemsService } from "./services/mmnt-items";
import { MarketService } from "./services/market";

const V1 = "/v1/";

// Init all routes, setup middlewares and dependencies
const initApp = (
  momentsService: MomentsService,
  mmntItemsService: MmntItemsService,
  marketService: MarketService
) => {
  const app = express();

  // @ts-ignore
  app.use(cors());
  app.use(json());
  app.use(urlencoded({ extended: false }));
  app.use(V1, initMomentsRouter(momentsService));
  app.use(V1, initMmntItemsRouter(mmntItemsService));
  app.use(V1, initMarketRouter(marketService));

  if (process.env.IS_HEROKU) {
    // Serve React static site using Express when deployed to Heroku.
    app.use(express.static(path.resolve(__dirname, "../../web/build")));
    app.get("*", function (req, res) {
      res.sendFile(path.resolve(__dirname, "../../web/build/index.html"));
    });
  }

  app.all("*", async (req: Request, res: Response) => {
    return res.sendStatus(404);
  });

  return app;
};

export default initApp;
