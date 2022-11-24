declare global {
  namespace NodeJS {
    interface ProcessEnv {
      PORT?: string;
    }
  }
}

export {};
