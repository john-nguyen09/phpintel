export function intToBytes(value: number): string {
    let bytes = [4];
    let i = 4;

    do {
        bytes[--i] = value & 255;
        value = value >> 8;
    } while (i > 0);

    let str = '';
    for (let i = 0; i < bytes.length; i++) {
        str += String.fromCharCode(bytes[i]);
    }
    
    return str;
}

export function multipleIntToBytes(...values: number[]): string {
    let str = '';

    for (let value of values) {
        str += intToBytes(value);
    }

    return str;
}

export function bytesToInts(str: string): number[] {
    let bytes: number[] = [];
    let intSize = 4;
    let results: number[] = [];

    for (let i = 0; i < str.length; i++) {
        bytes.push(str.charCodeAt(i));
    }

    for (let i = 0; i < bytes.length; i += intSize) {
        let val = 0;

        for (let j = 0; j < intSize; j++) {
            val += bytes[i + j];
            if (j < intSize - 1) {
                val = val << 8;
            }
        }

        results.push(val);
    }

    return results;
}