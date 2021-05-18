import express, { Request, Response, Router } from "express";
import { MmntItemsService } from "../services/mmnt-items";
import { body } from "express-validator";
import { validateRequest } from "../middlewares/validate-request";

function initMmntItemsRouter(mmntItemsService: MmntItemsService): Router {
  const router = express.Router();

  router.post(
    "/mmnt-items/mint",
    [body("recipient").exists(), body("typeID").isInt()],
    validateRequest,
    async (req: Request, res: Response) => {
      const { recipient, typeID } = req.body;
      const tx = await mmntItemsService.mint(recipient, typeID);
      return res.send({
        transaction: tx,
      });
    }
  );

  router.post("/mmnt-items/setup", async (req: Request, res: Response) => {
    const transaction = await mmntItemsService.setupAccount();
    return res.send({
      transaction,
    });
  });

  router.post(
    "/mmnt-items/transfer",
    [body("recipient").exists(), body("itemID").isInt()],
    validateRequest,
    async (req: Request, res: Response) => {
      const { recipient, itemID } = req.body;
      const tx = await mmntItemsService.transfer(recipient, itemID);
      return res.send({
        transaction: tx,
      });
    }
  );

  router.get(
    "/mmnt-items/collection/:account",
    async (req: Request, res: Response) => {
      const collection = await mmntItemsService.getCollectionIds(
        req.params.account
      );
      return res.send({
        collection,
      });
    }
  );

  router.get(
    "/mmnt-items/item/:itemID",
    async (req: Request, res: Response) => {
      const item = await mmntItemsService.getMmntItemType(
        parseInt(req.params.itemID)
      );
      return res.send({
        item,
      });
    }
  );

  router.get("/mmnt-items/supply", async (req: Request, res: Response) => {
    const supply = await mmntItemsService.getSupply();
    return res.send({
      supply,
    });
  });

  return router;
}

export default initMmntItemsRouter;
