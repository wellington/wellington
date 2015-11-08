#ifndef SASS_EMITTER_H
#define SASS_EMITTER_H

#include <string>
#include "sass/base.h"
#include "source_map.hpp"
#include "ast_fwd_decl.hpp"

namespace Sass {
  class Context;

  class Emitter {

    public:
      Emitter(Context* ctx);
      virtual ~Emitter() { }

    protected:
      OutputBuffer wbuf;
    public:
      const std::string& buffer(void) { return wbuf.buffer; }
      const SourceMap smap(void) { return wbuf.smap; }
      const OutputBuffer output(void) { return wbuf; }
      // proxy methods for source maps
      void add_source_index(size_t idx);
      void set_filename(const std::string& str);
      void add_open_mapping(AST_Node* node);
      void add_close_mapping(AST_Node* node);
      std::string render_srcmap(Context &ctx);
      ParserState remap(const ParserState& pstate);

    public:
      Context* ctx;
      size_t indentation;
      size_t scheduled_space;
      size_t scheduled_linefeed;
      bool scheduled_delimiter;

    public:
      // output strings different in comments
      bool in_comment;
      // selector list does not get linefeeds
      bool in_wrapped;
      // lists always get a space after delimiter
      bool in_media_block;
      // nested list must not have parentheses
      bool in_declaration;
      // nested lists need parentheses
      bool in_space_array;
      bool in_comma_array;
      // list separators don't get compressed in @debug
      bool in_debug;

    public:
      // return buffer as std::string
      std::string get_buffer(void);
      // flush scheduled space/linefeed
      Sass_Output_Style output_style(void);
      // add outstanding linefeed
      void finalize(void);
      // flush scheduled space/linefeed
      void flush_schedules(void);
      // prepend some text or token to the buffer
      void prepend_string(const std::string& text);
      void prepend_output(const OutputBuffer& out);
      // append some text or token to the buffer
      void append_string(const std::string& text);
      // append some white-space only text
      void append_wspace(const std::string& text);
      // append some text or token to the buffer
      // this adds source-mappings for node start and end
      void append_token(const std::string& text, AST_Node* node);

    public: // syntax sugar
      void append_indentation();
      void append_optional_space(void);
      void append_mandatory_space(void);
      void append_special_linefeed(void);
      void append_optional_linefeed(void);
      void append_mandatory_linefeed(void);
      void append_scope_opener(AST_Node* node = 0);
      void append_scope_closer(AST_Node* node = 0);
      void append_comma_separator(void);
      void append_colon_separator(void);
      void append_delimiter(void);

  };

}

#endif
