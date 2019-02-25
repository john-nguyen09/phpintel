import { ImportTable } from './importTable';

export class TypeName {
    public static readonly BUILT_INS = [
        'boolean',
        'int',
        'float',
        'null',
        'string'
    ];

    public static readonly ALIASES: Map<string, string> = new Map<string, string>([
        ['bool', 'boolean'],
        ['double', 'float'],
        ['integer', 'int'],
    ]);

    public name: string;

    constructor(name: string) {
        this.name = name.trim();

        const alias = TypeName.ALIASES.get(this.name);
        if (alias !== undefined) {
            this.name = alias;
        }
    }

    public resolveReferenceToFqn(importTable: ImportTable) {
        if (this.isVariable()) {
            return;
        }

        this.name = importTable.getFqn(this.name);
    }

    public resolveDefinitionToFqn(importTable: ImportTable) {
        if (TypeName.isBuiltin(this.name) || TypeName.isFqn(this.name)) {
            return;
        }

        this.name = importTable.namespace.fqn + this.name;
    }

    public static isBuiltin(typeName: string) {
        return TypeName.BUILT_INS.indexOf(typeName) >= 0;
    }

    public static isFqn(typeName: string): boolean {
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

    public isVariable(): boolean {
        return this.name.startsWith('$');
    }
}
