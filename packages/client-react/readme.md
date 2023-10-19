# @teamkeel/client-react

Create a fully typed `useKeel()` hook from a generated Keel client.

## Install

```
npm i @teamkeel/client-react
```

## Usage

Create the `KeelProvider` and `useKeel` components by passing your generated APIClient into the `keel` function from this package.

N.B. [See here](https://docs.keel.so/apis/client) for documentation on generating a client 

```ts
import { APIClient } from "../keelClient";
import { keel } from "@teamkeel/client-react";
export const { KeelProvider, useKeel } = keel(APIClient);

```

Wrap your app with the exported `KeelProvider` and set the endpoint for you API. 

The endpoint is the base URL + the api name (if you haven't manually set an API name this is `api`). This url can be found in the Keel web console or in the output of `keel run` in your terminal.

```tsx

<KeelProvider endpoint="https://myproject.keelapps.xyz/api/">
	<App />
</KeelProvider>

```

Now you can use the typed `useKeel` hook in any component within your app. You can [read more here](https://docs.keel.so/apis/client) on how to use the client.

```ts
function MyComponent() {
	const keel = useKeel();
}
```