import {
  BaseUiInputResponse,
  InputElementImplementation,
  InputElement,
} from "../..";

type ElementDataType = string;

/**
 * Defines what type of datepicker should we use.
 * @default "dateTime"
 */
type DatePickerMode =
  /** Pick both a date and a time */
  | "dateTime"
  /** Pick only a date */
  | "date";

export type UiElementInputDatePicker = InputElement<
  ElementDataType,
  {
    mode?: DatePickerMode;
    /** Allows selecting only dates/datetimes past this given date. */
    min?: string;
    /** Allows selecting only dates/datetimes before this given date. */
    max?: string;
  }
>;

// The shape of the response over the API
export interface UiElementInputDatePickerApiResponse
  extends BaseUiInputResponse<"ui.input.datePicker", ElementDataType> {
  mode?: DatePickerMode;
  min?: string;
  max?: string;
}

export const datePickerInput: InputElementImplementation<
  ElementDataType,
  UiElementInputDatePicker,
  UiElementInputDatePickerApiResponse
> = (name, options) => {
  return {
    __type: "input",
    uiConfig: {
      __type: "ui.input.datePicker",
      name,
      label: options?.label || name,
      optional: options?.optional || false,
      disabled: options?.disabled || false,
      helpText: options?.helpText,
      defaultValue: options?.defaultValue,
      mode: options?.mode,
      min: options?.min,
      max: options?.max,
    } satisfies UiElementInputDatePickerApiResponse,
    validate: options?.validate,
    onLeave: options?.onLeave,
    getData: (x: ElementDataType) => x,
  };
};
