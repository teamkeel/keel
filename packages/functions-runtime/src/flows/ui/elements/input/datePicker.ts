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
type DatepickerMode =
  /** Pick both a date and a time */
  | "dateTime"
  /** Pick only a date */
  | "date";

export type UiElementInputDatePicker = InputElement<
  ElementDataType,
  {
    mode?: DatepickerMode;
    /** Allows selecting only past dates/datetimes. */
    pastOnly?: boolean;
  }
>;

// The shape of the response over the API
export interface UiElementInputDatePickerApiResponse
  extends BaseUiInputResponse<"ui.input.datePicker", ElementDataType> {
  mode?: DatepickerMode;
  pastOnly?: boolean;
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
      pastOnly: options?.pastOnly,
    } satisfies UiElementInputDatePickerApiResponse,
    validate: options?.validate,
    onLeave: options?.onLeave,
    getData: (x: ElementDataType) => x,
  };
};
