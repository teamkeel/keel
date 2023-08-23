import { Executor } from "./Executor.mjs";

export class SubscriberExecutor extends Executor {
  constructor(props) {
    props.apiBaseUrl = process.env.KEEL_TESTING_SUBSCRIBERS_URL;
    props.parseJsonResult = false;

    super(props);
  }
}
