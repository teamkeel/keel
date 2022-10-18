
export default function log(msg: string, silent: boolean) {
  if (!silent) {
    console.log(msg);
  }
}