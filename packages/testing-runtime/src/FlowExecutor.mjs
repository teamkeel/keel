import { Executor } from "./Executor.mjs";

export class FlowExecutor extends Executor {
  constructor(props) {
    props.apiBaseUrl = process.env.KEEL_TESTING_FLOWS_URL;
    props.parseJsonResult = false;

    super(props);
  }
}
