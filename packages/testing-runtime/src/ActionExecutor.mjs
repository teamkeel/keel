import { Executor } from "./Executor.mjs";

export class ActionExecutor extends Executor {
  constructor(props) {
    props.apiBaseUrl = process.env.KEEL_TESTING_ACTIONS_API_URL + "/json";
    props.parseJsonResult = true;

    super(props);
  }
}
