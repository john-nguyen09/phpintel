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
import { DocNode, ParamDocNode } from "../../docParser";

export class Function extends Symbol implements Consumer, DocBlockConsumer {
    public name: TypeName = null;
    public parameters: Parameter[] = [];
    public scopeVar: Scope = new Scope();
    public typeAggregate: TypeComposite = new TypeComposite();
    public description: string = '';

    private docParamTypes: {[key: string]: TypeName} = {};

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
            } else if (returnSymbol instanceof Expression) {
                this.typeAggregate.push(returnSymbol.type);
            }

            return true;
        } else if (other instanceof SimpleVariable) {
            this.scopeVar.set(other);

            return true;
        }

        return false;
    }

    consumeDocBlock(doc: DocBlock) {
        let docAst = doc.docAst;

        this.description = docAst.summary;

        for (let docNode of docAst.body) {
            if (DocBlock.isType<ParamDocNode>(docNode, 'param')) {
                this.docParamTypes['$' + docNode.name] = new TypeName(docNode.type.name);
            }
        }
    }

    get types(): TypeName[] {
        return this.typeAggregate.types;
    }
}