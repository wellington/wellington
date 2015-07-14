#ifndef SASS_EVAL_H
#define SASS_EVAL_H

#include <iostream>
#include "context.hpp"
#include "listize.hpp"
#include "operation.hpp"

namespace Sass {
  using namespace std;

  class Expand;
  class Context;
  class Listize;

  class Eval : public Operation_CRTP<Expression*, Eval> {

   private:
    Expression* fallback_impl(AST_Node* n);

   public:
    Expand&  exp;
    Context& ctx;
    Listize  listize;
    Eval(Expand& exp);
    virtual ~Eval();

    Env* environment();
    Context& context();
    Selector_List* selector();
    Backtrace* backtrace();

    using Operation<Expression*>::operator();

    // for evaluating function bodies
    Expression* operator()(Block*);
    Expression* operator()(Assignment*);
    Expression* operator()(If*);
    Expression* operator()(For*);
    Expression* operator()(Each*);
    Expression* operator()(While*);
    Expression* operator()(Return*);
    Expression* operator()(Warning*);
    Expression* operator()(Error*);
    Expression* operator()(Debug*);

    Expression* operator()(List*);
    Expression* operator()(Map*);
    Expression* operator()(Binary_Expression*);
    Expression* operator()(Unary_Expression*);
    Expression* operator()(Function_Call*);
    Expression* operator()(Function_Call_Schema*);
    Expression* operator()(Variable*);
    Expression* operator()(Textual*);
    Expression* operator()(Number*);
    Expression* operator()(Boolean*);
    Expression* operator()(String_Schema*);
    Expression* operator()(String_Quoted*);
    Expression* operator()(String_Constant*);
    // Expression* operator()(Selector_List*);
    Expression* operator()(Media_Query*);
    Expression* operator()(Media_Query_Expression*);
    Expression* operator()(At_Root_Expression*);
    Expression* operator()(Supports_Query*);
    Expression* operator()(Supports_Condition*);
    Expression* operator()(Null*);
    Expression* operator()(Argument*);
    Expression* operator()(Arguments*);
    Expression* operator()(Comment*);

    // these will return selectors
    Selector_List* operator()(Selector_List*);
    Selector_List* operator()(Complex_Selector*);
    Attribute_Selector* operator()(Attribute_Selector*);
    // they don't have any specific implementatio (yet)
    Type_Selector* operator()(Type_Selector* s) { return s; };
    Pseudo_Selector* operator()(Pseudo_Selector* s) { return s; };
    Wrapped_Selector* operator()(Wrapped_Selector* s) { return s; };
    Selector_Qualifier* operator()(Selector_Qualifier* s) { return s; };
    Selector_Placeholder* operator()(Selector_Placeholder* s) { return s; };
    // actual evaluated selectors
    Selector_List* operator()(Selector_Schema*);
    Expression* operator()(Parent_Selector*);

    template <typename U>
    Expression* fallback(U x) { return fallback_impl(x); }

  private:
    string interpolation(Expression* s);

  };

  Expression* cval_to_astnode(Sass_Value* v, Context& ctx, Backtrace* backtrace, ParserState pstate = ParserState("[AST]"));

  bool eq(Expression*, Expression*, Context&);
  bool lt(Expression*, Expression*, Context&);
}

#endif
