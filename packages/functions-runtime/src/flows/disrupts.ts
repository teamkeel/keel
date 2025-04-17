// This is a special type that is thrown to disrupt the execution of a flow
abstract class FlowDisrupt {
  protected constructor() {}
}

export class UIRenderDisrupt extends FlowDisrupt {
  constructor(public readonly stepId: string, public readonly page: string) {
    super();
  }
}

export class StepErrorDisrupt extends FlowDisrupt {
  constructor(public readonly errorMessage: string) {
    super();
  }
}

export class StepCompletedDisrupt extends FlowDisrupt {
  constructor() {
    super();
  }
}
