import { Symbol, Consumer, DocBlockConsumer } from "../symbol";
import { SimpleVariable } from "./simpleVariable";
import { DocBlock } from "../docBlock";
import { GlobalDocNode, toTypeName } from "../../util/docParser";
import { Variable } from "./variable";
import { ScopeVar } from "./scopeVar";
import { inspect } from "util";

export class GlobalVariable extends Symbol implements Consumer, DocBlockConsumer {
    public variables: Variable[] = [];

    private _isDefinition: boolean = false;
    private noNamedGlobalDocNodes: GlobalDocNode[] = [];
    private globalDocNodes: Map<string, GlobalDocNode> = new Map<string, GlobalDocNode>();
    private scopeVar: ScopeVar | null = null;

    public consume(other: Symbol): boolean {
        if (other instanceof SimpleVariable) {
            let globalDocNode: GlobalDocNode | null = null;

            if (this.globalDocNodes.has(other.name)) {
                globalDocNode = this.globalDocNodes.get(other.name) || null;
            } else {
                const noNamedGlobalDocNode = this.noNamedGlobalDocNodes.shift();
                if (noNamedGlobalDocNode !== undefined) {
                    globalDocNode = noNamedGlobalDocNode;
                }
            }

            const variable: Variable = new Variable(other.name, other.type);
            if (globalDocNode !== null) {
                variable.type.push(toTypeName(globalDocNode.type));
            }

            this.variables.push(variable);
            if (this.scopeVar !== null) {
                this.scopeVar.addGlobalVariableName(variable.name);
            }
        }

        return true;
    }

    public consumeDocBlock(doc: DocBlock) {
        const docAst = doc.docAst;
        if (docAst.kind == 'doc') {
            const globalDocNodes = doc.getNodes<GlobalDocNode>('global');
            this._isDefinition = globalDocNodes.length > 0;

            for (const globalDocNode of globalDocNodes) {
                if (globalDocNode.variable !== null) {
                    this.globalDocNodes.set('$' + globalDocNode.variable, globalDocNode);
                } else {
                    this.noNamedGlobalDocNodes.push(globalDocNode);
                }
            }
        }
    }

    public setScopeVar(scopeVar: ScopeVar) {
        this.scopeVar = scopeVar;
    }

    public assignExtraTypeForVariables() {
        if (this.scopeVar === null) {
            return;
        }

        for (let i = 0; i < this.variables.length; i++) {
            const types = this.scopeVar.getType(this.variables[i].name);

            for (const type of types.types) {
                this.variables[i].type.push(type);
            }
        }
    }

    get isDefinition(): boolean {
        return this._isDefinition;
    }
}