import { ImportTable } from './importTable';

export class TypeName {
    public static readonly BUILT_INS = [
        'boolean',
        'int',
        'float',
        'null',
        'string'
    ];

    public static readonly ALIASES: {[alias: string]: string} = {
        'bool': 'boolean',
        'double': 'float',
        'integer': 'int'
    };

    private name: string;
    public isArray: boolean = false;

    constructor(name: string, isArray?: boolean) {
        this.name = name;

        if (this.name in TypeName.ALIASES) {
            this.name = TypeName.ALIASES[this.name];
        }

        if (isArray !== undefined) {
            this.isArray = isArray;
        }
    }

    public resolveToFullyQualified(importTable: ImportTable) {
        this.name = importTable.getFqn(this.name);
    }

    public static isFullyQualifiedName(typeName: string): boolean {
        if (typeName == '') {
            return true;
        }

        if (TypeName.BUILT_INS.indexOf(typeName) >= 0) {
            return true;
        }

        return typeName.indexOf('\\') === 0;
    }

    public getName(): string {
        return this.name;
    }

    public toString(): string {
        return this.name + (this.isArray ? '[]' : '');
    }

    public isSameAs(other: TypeName): boolean {
        if (other == undefined) {
            return false;
        }

        return (this.isArray == other.isArray) && this.name == other.name
    }

    public isEmptyName(): boolean {
        return this.name == '';
    }
}
