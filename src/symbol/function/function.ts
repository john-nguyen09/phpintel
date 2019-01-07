import { Symbol, Consumer, DocBlockConsumer, NamedSymbol, Locatable, NameResolvable } from "../symbol";
import { FunctionHeader } from "./functionHeader";
import { Parameter } from "../variable/parameter";
import { ScopeVar } from "../variable/scopeVar";
import { Return } from "../type/return";
import { Variable } from "../variable/variable";
import { Expression } from "../type/expression";
import { SimpleVariable } from "../variable/simpleVariable";
import { TypeComposite } from "../../type/composite";
import { TypeName } from "../../type/name";
import { DocBlock } from "../docBlock";
import { DocNodeKind, toTypeName } from "../../util/docParser";
import { VariableAssignment } from "../variable/varibleAssignment";
import { FieldGetter } from "../fieldGetter";
import { ImportTable } from "../../type/importTable";
import { Location } from "../meta/location";

export class Function extends Symbol implements
    Consumer,
    DocBlockConsumer,
    FieldGetter,
    NamedSymbol,
    Locatable,
    NameResolvable {
    public name: TypeName;
    public location: Location = new Location();
    public parameters: Parameter[] = [];
    public scopeVar: ScopeVar = new ScopeVar();
    public typeAggregate: TypeComposite = new TypeComposite();
    public description: string = '';

    private docParamTypes: { [key: string]: TypeName } = {};

    consume(other: Symbol) {
        if (other instanceof Parameter) {
            if (other.name in this.docParamTypes) {
                other.type.push(this.docParamTypes[other.name]);
            }

            this.parameters.push(other);
            this.scopeVar.consume(other);

            return true;
        } else if (other instanceof FunctionHeader) {
            this.name = other.name;

            return true;
        } else if (other instanceof Return) {
            let returnSymbol = other.returnSymbol;

            if (returnSymbol instanceof Variable) {
                let types = this.scopeVar.getType(returnSymbol.name).types;

                for (let type of types) {
                    this.typeAggregate.push(type);
                }
            } else if (
                returnSymbol instanceof Expression &&
                returnSymbol.type != undefined
            ) {
                this.typeAggregate.push(returnSymbol.type);
            }

            return true;
        } else if (other instanceof VariableAssignment) {
            if (other.variable != undefined) {
                this.scopeVar.set(other.variable);
            }

            return true;
        } else if (other instanceof SimpleVariable) {
            return true;
        }

        return false;
    }

    consumeDocBlock(doc: DocBlock) {
        let docAst = doc.docAst;

        this.description = docAst.summary;

        for (let docNode of docAst.body) {
            if (docNode.kind == DocNodeKind.Param) {
                let typeName = toTypeName(docNode.type);

                if (typeName != null) {
                    this.docParamTypes['$' + docNode.name] = typeName;
                }
            }
        }
    }

    get types(): TypeName[] {
        return this.typeAggregate.types;
    }

    getFields(): string[] {
        return [
            'name', 'parameters', 'scopeVar', 'types', 'description'
        ];
    }

    public getName(): string {
        return this.name.toString();
    }

    public resolveName(importTable: ImportTable): void {
        for (let param of this.parameters) {
            for (let type of param.type.types) {
                type.resolveToFullyQualified(importTable);
            }
        }

        this.name.resolveToFullyQualified(importTable);

        for (let type of this.typeAggregate.types) {
            type.resolveToFullyQualified(importTable);
        }
    }
}