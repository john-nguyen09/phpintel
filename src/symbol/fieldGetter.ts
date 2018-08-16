export interface FieldGetter {
    getFields(): string[];
}

export function isFieldGetter(object: Object): object is (Object & FieldGetter) {
    return 'getFields' in object &&
        typeof (<any>object).getFields == 'function';
}