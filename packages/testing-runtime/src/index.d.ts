// See https://vitest.dev/guide/extending-matchers.html for docs
// on typing custom matchers
import type { Assertion, AsymmetricMatchersContaining } from "vitest";

interface ActionError {
  code: string;
  message: string;
}

interface CustomMatchers<R = unknown> {
  toHaveAuthorizationError(): void;
  toHaveAuthenticationError(): void;
  toHaveError(err: Partial<ActionError>): void;
}

declare module "vitest" {
  interface Assertion<T = any> extends CustomMatchers<T> {}
  interface AsymmetricMatchersContaining extends CustomMatchers {}
}

// Flow Status Types
export type FlowStatus =
  | "NEW"
  | "RUNNING"
  | "AWAITING_INPUT"
  | "FAILED"
  | "COMPLETED"
  | "CANCELLED";

// Step Types
export type StepType = "FUNCTION" | "UI" | "COMPLETE";

// Step Status Types
export type StepStatus =
  | "NEW"
  | "PENDING"
  | "FAILED"
  | "COMPLETED"
  | "CANCELLED";

// Stage Configuration
export interface FlowStage {
  key: string;
  name: string;
  description: string;
}

// Flow Configuration
export interface FlowConfig {
  title: string;
  description?: string;
  stages?: FlowStage[];
}

// UI Action Configuration
export interface UIAction {
  label: string;
  mode: "primary" | "secondary";
  value: string;
}

// UI Input Elements
export interface UITextInput {
  __type: "ui.input.text";
  label: string;
  name: string;
  disabled?: boolean;
  optional?: boolean;
  defaultValue?: string;
  placeholder?: string;
  validationError?: string;
}

export interface UINumberInput {
  __type: "ui.input.number";
  label: string;
  name: string;
  disabled?: boolean;
  optional?: boolean;
  defaultValue?: number;
  validationError?: string;
}

export interface UIBooleanInput {
  __type: "ui.input.boolean";
  label: string;
  name: string;
  disabled?: boolean;
  optional?: boolean;
  mode: "checkbox";
  validationError?: string;
}

// UI Display Elements
export interface UIDivider {
  __type: "ui.display.divider";
}

export interface UIMarkdown {
  __type: "ui.display.markdown";
  content: string;
}

export interface UIGrid {
  __type: "ui.display.grid";
  data: Array<{ title: string; [key: string]: any }>;
}

// UI Complete Element
export interface UIComplete {
  __type: "ui.complete";
  title: string;
  description?: string;
  stage?: string;
  content: UIElement[];
}

// UI Page Element
export interface UIPage {
  __type: "ui.page";
  title: string;
  description?: string;
  content: UIElement[];
  actions?: UIAction[];
  hasValidationErrors: boolean;
  validationError?: string;
}

// Union type for all UI elements
export type UIElement =
  | UITextInput
  | UINumberInput
  | UIBooleanInput
  | UIDivider
  | UIMarkdown
  | UIGrid;

// Union type for UI configurations
export type UIConfig = UIPage | UIComplete;

declare class FlowExecutor<Input = {}> {
  withIdentity(identity: sdk.Identity): FlowExecutor<Input>;
  withAuthToken(token: string): FlowExecutor<Input>;
  start(inputs: Input): Promise<FlowRun<Input>>;
  get(id: string): Promise<FlowRun<Input>>;
  cancel(id: string): Promise<FlowRun<Input>>;
  putStepValues(
    id: string,
    stepId: string,
    values: Record<string, any>,
    action?: string
  ): Promise<FlowRun<Input>>;
  untilAwaitingInput(id: string, timeout?: number): Promise<FlowRun<Input>>;
  untilFinished(id: string, timeout?: number): Promise<FlowRun<Input>>;
}

// Step Definition
export interface FlowStep {
  id: string;
  name: string;
  runId: string;
  stage: string | null;
  status: StepStatus;
  type: StepType;
  value: any;
  error: string | null;
  startTime: string | null;
  endTime: string | null;
  createdAt: string;
  updatedAt: string;
  ui: UIConfig | null;
}

// Flow Run Definition
export interface FlowRun<Input = {}> {
  id: string;
  traceId: string;
  status: FlowStatus;
  name: string;
  startedBy: Date;
  input: Input | {};
  data: any;
  steps: FlowStep[];
  createdAt: Date;
  updatedAt: Date;
  config: FlowConfig;
}

// Step Values Request
export interface PutStepValuesRequest {
  name: string;
  runId: string;
  stepId: string;
  values: Record<string, any>;
  token: string;
  action?: string | null;
}

// Flow Start Request
export interface StartFlowRequest {
  name: string;
  token: string;
  body: Record<string, any>;
}

// Flow Get Request
export interface GetFlowRequest {
  name: string;
  id: string;
  token: string;
}
