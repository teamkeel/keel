// This is a special type that is thrown to disrupt the execution of a flow
abstract class FlowDisrupt {
  protected constructor() {}
}

export class UIRenderDisrupt extends FlowDisrupt {
  constructor(
    public readonly stepId: string,
    public readonly contents: any
  ) {
    super();
  }
}

export class StepErrorDisrupt extends FlowDisrupt {
  constructor(public readonly message: string) {
    super();
  }
}

export class StepCreatedDisrupt extends FlowDisrupt {
  constructor(
    public readonly executeAfter?: Date,
  ) {
    super();
  }
}

export class ExhuastedRetriesDisrupt extends FlowDisrupt {
  constructor() {
    super();
  }
}
