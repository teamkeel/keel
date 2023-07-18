# @teamkeel/client-react

Typed useKeel Hook from a generated Keel client

## Install

```
npm i @teamkeel/client-react
```

## Usage

Export the `KeelProvider` and `useKeel` components based on your generated client.

N.B. [See here](https://docs.keel.so/apis/client) for documentation on generating a client 

```ts
import { APIClient } from "../keelClient";
export const { KeelProvider, useKeel } = keel(APIClient);

```

Wrap your app with the exported `KeelProvider`

```tsx

<KeelProvider endpoint="https://myproject.keelapps.xyz/api/">
	<App />
</KeelProvider>

```

Now you can use the typed `useKeel` hook in any component within your app

```ts
function MyComponent() {
	const keel = useKeel();
}
```