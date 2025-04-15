
class UI {
  constructor() {
    this.inputs = {
        text: (name, options) => {
            return {
                _type: "input",
                name: name,
                valueType: "string",
                uiConfig: {
                    __type: "ui.input.text",
                    name: name,
                    label: options?.label || name,
                    defaultValue: options?.defaultValue,
                    optional: options?.optional,
                    maxLength: options?.maxLength
                }
            };
        },
        toggle: (name, options) => {
            return {
                _type: "input",
                name: name,
                valueType: "boolean",
                uiConfig: {
                    __type: "ui.input.toggle",
                    name: name,
                    label: options?.label || name,
                    defaultValue: options?.defaultValue,
                    optional: options?.optional,
                }
            };
        }
    };
    
    this.display = {
        divider: () => {
            return {
                _type: "display",
                uiConfig: {
                    __type: "ui.display.divider"
                }
            };
        }
    };
}
  

  async page(options) {
    console.log(options);

    // Get this step from the database and determine next move:
    //  - If the step is RUNNING and there is no data, then return flow UI structure.  We are still waiting on UI data.
    //  - If the step is RUNNING and there is data, then run the validation functions. If these all pass, then update the step to COMPLETED.
    //  - If the step is COMPLETED, then return the data.

    return {};
  }

 
}

module.exports = { UI };
