import React, { createContext, useContext, useEffect, useRef } from "react";

interface KeelContextType<T> {
  client: T;
}

const KeelContext = createContext<KeelContextType<any>>({
  client: null,
});

interface KeelProviderProps<
  T extends new (...args: any[]) => any,
  U = Omit<ConstructorParameters<T>[0], "endpoint">,
> {
  /**
   * The base URL for the client.
   */
  baseUrl: string;
  /**
   * Additional config options for the client.
   */
  config?: U;
  children: React.ReactNode;
}

export const keel = <T extends new (...args: any[]) => any>(Client: T) => {
  function KeelProvider<T extends new (...args: any[]) => any>({
    baseUrl,
    config,
    children,
  }: KeelProviderProps<T>) {
    if (typeof Client !== "function") {
      throw new Error("Client must be a Keel class");
    }

    const clientConstructor = Client as new (args: any) => any;
    const clientArgs = { baseUrl, ...config };
    const clientRef = useRef(new clientConstructor(clientArgs));

    const client = clientRef.current;

    useEffect(() => {
      client.client.setBaseUrl(baseUrl);
    }, [baseUrl, client]);

    return (
      <KeelContext.Provider value={{ client }}>{children}</KeelContext.Provider>
    );
  }

  return {
    KeelProvider: KeelProvider<T>,
    useKeel: useKeel<T>,
  };
};

function useKeel<T extends new (...args: any) => any>() {
  const keelContext = useContext<KeelContextType<InstanceType<T>>>(KeelContext);

  if (!keelContext) {
    throw new Error("useKeel must be used within a KeelProvider");
  }

  return keelContext.client;
}
