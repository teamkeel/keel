# @teamkeel/client-react-query

Typed React-Query Hook from a generated Keel client

## Install

N.B. Requires `@teamtkeel/client-react` to be installed and setup first.

```
npm i @teamkeel/client-react-query
```

## Usage

Export `useKeelQuery` and `useKeelMutation` by passing in a `useKeel` hook based on your api client.

```ts
// Follow @teamtkeel/client-react setup instructions
import { APIClient } from "../keelClient";
import { keel } from "@teamkeel/client-react";

import { keelQuery } from "@teamkeel/client-react-query";

export const { KeelProvider, useKeel } = keel(APIClient);
export const { useKeelQuery, useKeelMutation } = keelQuery(useKeel);

```

You can then use these hooks in your components. Inputs are fully typed based on the action name.

```tsx

const query = useKeelQuery("actionName", actionInputs, queryOptions);

const mutation = useKeelMutation("actionName", mutationOptions);

```