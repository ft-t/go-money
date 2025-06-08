export function uint8ArrayToBase64(array: Uint8Array): string {
  let res = '';

  for (let i = 0; i < array.byteLength; i++) {
    res += String.fromCharCode(array[i]);
  }

  return window.btoa(res);
}
