import { TypeName } from './name';

export class ImportTable {
    private namespace: string = '';
    private imports: {[key: string]: string} = {};

    constructor() { }

    public setNamespace(namespace: string) {
        this.namespace = namespace;
    }

    public import(fqn: string, alias?: string) {
        let parts = fqn.split('\\').filter((part) => {
            return part != '';
        });

        if (!alias) {
            alias = parts.pop();
        }

        this.imports[alias] = '\\' + parts.join('\\');
    }

    public getFqn(name: string) {
        if (TypeName.isFullyQualifiedName(name)) {
            return name;
        }

        let parts = name.split('\\');
        let alias = parts.shift();
        let namespace: string;

        if (alias in this.imports) {
            namespace = this.imports[alias];
        } else {
            namespace = this.namespace;
        }

        return namespace + (parts.length > 0 ? '\\' + parts.join('\\') : '');
    }
}
