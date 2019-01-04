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

    public name: string;

    constructor(name: string) {
        this.name = name.trim();

        if (this.name in TypeName.ALIASES) {
            this.name = TypeName.ALIASES[this.name];
        }
    }

    public resolveToFullyQualified(importTable: ImportTable) {
        this.name = importTable.getFqn(this.name);
    }

    public static isBuiltin(typeName: string) {
        return TypeName.BUILT_INS.indexOf(typeName) >= 0;
    }

    public static isFullyQualifiedName(typeName: string): boolean {
        if (typeName == '') {
            return true;
        }

        return typeName.indexOf('\\') === 0;
    }

    public getQualified(importTable: ImportTable): string {
        return importTable.getQualified(this.name);
    }

    public toString(): string {
        return this.name;
    }

    public isEmptyName(): boolean {
        return this.name == '';
    }
}
