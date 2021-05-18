import express, { Request, Response, Router } from "express";
import { MomentsService } from "../services/moment";
import { body } from "express-validator";
import { validateRequest } from "../middlewares/validate-request";

function initMomentsRouter(momentsService: MomentsService): Router {
  const router = express.Router();

  router.post(
    "/moments/mint",
    [body("recipient").exists(), body("amount").isDecimal()],
    validateRequest,
    async (req: Request, res: Response) => {
      const { recipient, amount } = req.body;

      const transaction = await momentsService.mint(recipient, amount);
      return res.send({
        transaction,
      });
    }
  );

  router.post("/moments/setup", async (req: Request, res: Response) => {
    const transaction = await momentsService.setupAccount();
    return res.send({
      transaction,
    });
  });

  router.post(
    "/moments/burn",
    [
      body("amount").isInt({
        gt: 0,
      }),
    ],
    validateRequest,
    async (req: Request, res: Response) => {
      const { amount } = req.body;
      const transaction = await momentsService.burn(amount);
      return res.send({
        transaction,
      });
    }
  );

  router.post(
    "/moments/transfer",
    [
      body("recipient").exists(),
      body("amount").isInt({
        gt: 0,
      }),
    ],
    validateRequest,
    async (req: Request, res: Response) => {
      const { recipient, amount } = req.body;
      const transaction = await momentsService.transfer(recipient, amount);
      return res.send({
        transaction,
      });
    }
  );

  router.get(
    "/moments/balance/:account",
    async (req: Request, res: Response) => {
      const balance = await momentsService.getBalance(req.params.account);
      return res.send({
        balance,
      });
    }
  );

  router.get("/moments/supply", async (req: Request, res: Response) => {
    const supply = await momentsService.getSupply();
    return res.send({
      supply,
    });
  });

  return router;
}

export default initMomentsRouter;
