import { Executor } from "./Executor.mjs";

export class JobExecutor extends Executor {
  constructor(props) {
    props.apiBaseUrl = process.env.KEEL_TESTING_JOBS_URL;
    props.parseJsonResult = false;

    super(props);
  }
}
