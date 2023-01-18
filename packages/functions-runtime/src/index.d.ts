export type IDWhereCondition = {
  equals?: string;
  oneOf?: string[];
};

export type StringWhereCondition = {
  startsWith?: string;
  endsWith?: string;
  oneOf?: string[];
  contains?: string;
  notEquals?: string;
  equals?: string;
};

export type BooleanWhereCondition = {
  equals?: boolean;
  notEquals?: boolean;
};

export type NumberWhereCondition = {
  greaterThan?: number;
  greaterThanOrEquals?: number;
  lessThan?: number;
  lessThanOrEquals?: number;
  equals?: number;
  notEquals?: number;
};

export type DateWhereCondition = {
  equals?: Date;
  before?: Date;
  onOrBefore?: Date;
  after?: Date;
  onOrAfter?: Date;
};
