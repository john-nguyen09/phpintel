import { Class } from "../symbol/class/class";

export default {
    type: 'symbol-encoding',
    encode: (symbol: Symbol): Buffer => {
        throw new Error(`${typeof symbol} encoding is not defined`);
    },
    decode: (buffer: Buffer): Symbol => {
        throw new Error(`Invalue buffer`);
    },
    buffer: false
} as Level.Encoding;

let classEncoding =  {
    encode: (symbol: Class): Buffer => {
        let buffer = Buffer.alloc(0);

        // buffer.write(symbol.name.)

        return buffer;
    }
}