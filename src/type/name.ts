import { ImportTable } from './importTable';

export class TypeName {
    public static readonly BUILT_INS = [
        'boolean',
        'int',
        'float',
        'null',
        'string'
    ];

    private name: string;
    private isArray: boolean = false;

    constructor(name: string) {
        let indexOfBox = name.lastIndexOf('[]');

        if ((name.length - indexOfBox) == '[]'.length) {
            this.isArray = true;
        }

        if (this.isArray) {
            this.name = name.substr(0, indexOfBox);
        } else {
            this.name = name;
        }
    }

    public resolveToFullyQualified(importTable: ImportTable) {
        this.name = importTable.getFqn(this.name);
    }

    public static isFullyQualifiedName(typeName: string): boolean {
        if (TypeName.BUILT_INS.indexOf(typeName) >= 0) {
            return true;
        }

        return typeName.indexOf('\\') === 0;
    }

    public toString(): string {
        return this.name + (this.isArray ? '[]' : '');
    }

    public isSameAs(other: TypeName): boolean {
        return (this.isArray == other.isArray) && this.name == other.name
    }
}
