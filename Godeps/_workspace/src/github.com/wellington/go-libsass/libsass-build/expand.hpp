#ifndef SASS_EXPAND_H
#define SASS_EXPAND_H

#include <map>
#include <vector>
#include <iostream>

#include "ast.hpp"
#include "eval.hpp"
#include "operation.hpp"
#include "environment.hpp"

namespace Sass {
  using namespace std;

  class Listize;
  class Context;
  class Eval;
  typedef Environment<AST_Node*> Env;
  struct Backtrace;

  class Expand : public Operation_CRTP<Statement*, Expand> {
  public:

    Env* environment();
    Context& context();
    Selector_List* selector();
    Backtrace* backtrace();

    Context&          ctx;
    Eval              eval;

    // it's easier to work with vectors
    vector<Env*>      env_stack;
    vector<Block*>    block_stack;
    vector<String*>   property_stack;
    vector<Selector_List*> selector_stack;
    vector<Backtrace*>backtrace_stack;
    bool              in_keyframes;

    Statement* fallback_impl(AST_Node* n);

  public:
    Expand(Context&, Env*, Backtrace*);
    virtual ~Expand() { }

    using Operation<Statement*>::operator();

    Statement* operator()(Block*);
    Statement* operator()(Ruleset*);
    Statement* operator()(Propset*);
    Statement* operator()(Media_Block*);
    Statement* operator()(Supports_Block*);
    Statement* operator()(At_Root_Block*);
    Statement* operator()(At_Rule*);
    Statement* operator()(Declaration*);
    Statement* operator()(Assignment*);
    Statement* operator()(Import*);
    Statement* operator()(Import_Stub*);
    Statement* operator()(Warning*);
    Statement* operator()(Error*);
    Statement* operator()(Debug*);
    Statement* operator()(Comment*);
    Statement* operator()(If*);
    Statement* operator()(For*);
    Statement* operator()(Each*);
    Statement* operator()(While*);
    Statement* operator()(Return*);
    Statement* operator()(Extension*);
    Statement* operator()(Definition*);
    Statement* operator()(Mixin_Call*);
    Statement* operator()(Content*);

    template <typename U>
    Statement* fallback(U x) { return fallback_impl(x); }

    void append_block(Block*);
  };

}

#endif
