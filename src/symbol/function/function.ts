import { Symbol, Consumer, DocBlockConsumer } from "../symbol";
import { FunctionHeader } from "./functionHeader";
import { Parameter } from "../variable/parameter";
import { Scope } from "../variable/scope";
import { Return } from "../type/return";
import { Variable } from "../variable/variable";
import { Expression } from "../type/expression";
import { SimpleVariable } from "../variable/simpleVariable";
import { TypeComposite } from "../../type/composite";
import { TypeName } from "../../type/name";
import { DocBlock } from "../docBlock";
import { ParamDocNode, DocNodeKind } from "../../docParser";
import { VariableAssignment } from "../variable/varibleAssignment";
import { FieldGetter } from "../../fieldGetter";

export class Function extends Symbol implements Consumer, DocBlockConsumer, FieldGetter {
    public name: TypeName;
    public parameters: Parameter[] = [];
    public scopeVar: Scope = new Scope();
    public typeAggregate: TypeComposite = new TypeComposite();
    public description: string = '';

    private docParamTypes: {[key: string]: TypeName} = {};

    consume(other: Symbol) {
        if (other instanceof Parameter) {
            if (other.name in this.docParamTypes) {
                other.type.push(this.docParamTypes[other.name]);

                if (this.doc != null) {
                    for (let type of other.type.types) {
                        type.resolveToFullyQualified(this.doc.importTable);
                    }
                }
            }

            this.parameters.push(other);
            this.scopeVar.consume(other);

            return true;
        } else if (other instanceof FunctionHeader) {
            this.name = other.name;

            if (this.doc != null) {
                this.name.resolveToFullyQualified(this.doc.importTable);
            }

            return true;
        } else if (other instanceof Return) {
            let returnSymbol = other.returnSymbol;

            if (returnSymbol instanceof Variable) {
                let types = this.scopeVar.getType(returnSymbol.name).types;

                for (let type of types) {
                    this.typeAggregate.push(type);
                }
            } else if (returnSymbol instanceof Expression) {
                if (this.doc != null) {
                    returnSymbol.type.resolveToFullyQualified(this.doc.importTable);
                }

                this.typeAggregate.push(returnSymbol.type);
            }

            return true;
        } else if (other instanceof VariableAssignment) {
            this.scopeVar.set(other.variable);

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
            if (DocBlock.isType<ParamDocNode>(docNode, DocNodeKind.Param)) {
                let type = docNode.type.name;

                if (docNode.type.fqn) {
                    type = '\\' + type;
                }

                let typeName = new TypeName(type);

                if (this.doc != null) {
                    typeName.resolveToFullyQualified(this.doc.importTable);
                }

                this.docParamTypes['$' + docNode.name] = typeName;
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
}