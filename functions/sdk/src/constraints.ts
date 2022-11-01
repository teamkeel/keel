// All of the different constraint types are unions of the underlying type
// or an object type which you can use to query by a set of permitted operators
// based on the type. e.g if you are querying a number field, then you can also perform number
// related operations on that field such as gte / lte etc
// Where the union resolves to the actual type such as string or number, this is equivalent
// to an equality check.

// sample query object:
// {
//   myStringField: "this is a string", // <== shorthand means "equal"
//   myNumberField: {
//     greaterThan: 10
//   }
//   myOtherNumberField: 10 // <== equality check
// }

export type Constraint<T> =
  | T
  | {
      equal?: T;
      notEqual?: T;

      // string constraints
      startsWith?: T extends String ? string : never;
      endsWith?: T extends String ? string : never;
      oneOf?: T extends String ? string[] : never;
      contains?: T extends String ? string : never;

      // number constraints
      greaterThan?: T extends Number ? number : never;
      greaterThanOrEqualTo?: T extends Number ? number : never;
      lessThan?: T extends Number ? number : never;
      lessThanOrEqualTo?: T extends Number ? number : never;

      // date constraints
      before?: T extends Date ? Date : never;
      onOrBefore?: T extends Date ? Date : never;
      after?: T extends Date ? Date : never;
      onOrAfter?: T extends Date ? Date : never;
    };
