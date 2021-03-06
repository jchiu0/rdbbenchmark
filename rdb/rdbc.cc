// This file is a subset of the C API from RocksDB. It should remain consistent.
// There will be another file which contains some extra routines that we find
// useful.
#include <cstddef>
#include <cstdlib>
#include <memory>
#include <string>

#include <rocksdb/cache.h>
#include <rocksdb/db.h>
#include <rocksdb/filter_policy.h>
#include <rocksdb/iterator.h>
#include <rocksdb/memtablerep.h>
#include <rocksdb/options.h>
#include <rocksdb/slice_transform.h>
#include <rocksdb/status.h>
#include <rocksdb/table.h>
#include <rocksdb/write_batch.h>

#include "_cgo_export.h"
#include "rdbc.h"

using rocksdb::DB;
using rocksdb::Options;
using rocksdb::Status;
using rocksdb::ReadOptions;
using rocksdb::Slice;
using rocksdb::WriteOptions;
using rocksdb::WriteBatch;
using rocksdb::Iterator;
using rocksdb::FilterPolicy;
using rocksdb::NewBloomFilterPolicy;
using rocksdb::Cache;
using rocksdb::NewLRUCache;
using rocksdb::BlockBasedTableOptions;
using rocksdb::SliceTransform;
using rocksdb::CompressionType;
using rocksdb::FlushOptions;

struct rocksdb_t { DB* rep; };
struct rocksdb_options_t { Options rep; };
struct rocksdb_readoptions_t {
  ReadOptions rep;
  Slice upper_bound; // stack variable to set pointer to in ReadOptions
};
struct rocksdb_writeoptions_t { WriteOptions rep; };
struct rocksdb_writebatch_t { WriteBatch rep; };
struct rocksdb_iterator_t { Iterator* rep; };
struct rocksdb_cache_t { std::shared_ptr<Cache> rep; };
struct rocksdb_block_based_table_options_t { BlockBasedTableOptions rep; };
struct rocksdb_flushoptions_t { FlushOptions rep; };

bool SaveError(char** errptr, const Status& s) {
  assert(errptr != nullptr);
  if (s.ok()) {
    return false;
  } else if (*errptr == nullptr) {
    *errptr = strdup(s.ToString().c_str());
  } else {
    // TODO(sanjay): Merge with existing error?
    // This is a bug if *errptr is not created by malloc()
    free(*errptr);
    *errptr = strdup(s.ToString().c_str());
  }
  return true;
}

static char* CopyString(const std::string& str) {
  char* result = reinterpret_cast<char*>(malloc(sizeof(char) * str.size()));
  memcpy(result, str.data(), sizeof(char) * str.size());
  return result;
}

//////////////////////////// rocksdb_t
rocksdb_t* rocksdb_open(
  const rocksdb_options_t* options,
  const char* name,
  char** errptr) {
  DB* db;
  if (SaveError(errptr, DB::Open(options->rep, std::string(name), &db))) {
    return nullptr;
  }
  rocksdb_t* result = new rocksdb_t;
  result->rep = db;
  return result;
}

rocksdb_t* rocksdb_open_for_read_only(
    const rocksdb_options_t* options,
    const char* name,
    unsigned char error_if_log_file_exist,
    char** errptr) {
  DB* db;
  if (SaveError(errptr, DB::OpenForReadOnly(options->rep, std::string(name), &db, error_if_log_file_exist))) {
    return nullptr;
  }
  rocksdb_t* result = new rocksdb_t;
  result->rep = db;
  return result;
}

void rocksdb_close(rocksdb_t* db) {
  delete db->rep;
  delete db;
}

char* rocksdb_get(
    rocksdb_t* db,
    const rocksdb_readoptions_t* options,
    const char* key, size_t keylen,
    size_t* vallen,
    char** errptr) {
  char* result = nullptr;
  std::string tmp;
  Status s = db->rep->Get(options->rep, Slice(key, keylen), &tmp);
  if (s.ok()) {
    *vallen = tmp.size();
    result = CopyString(tmp);
  } else {
    *vallen = 0;
    if (!s.IsNotFound()) {
      SaveError(errptr, s);
    }
  }
  return result;
}

void rocksdb_put(
    rocksdb_t* db,
    const rocksdb_writeoptions_t* options,
    const char* key, size_t keylen,
    const char* val, size_t vallen,
    char** errptr) {
  SaveError(errptr,
            db->rep->Put(options->rep, Slice(key, keylen), Slice(val, vallen)));
}

void rocksdb_delete(
    rocksdb_t* db,
    const rocksdb_writeoptions_t* options,
    const char* key, size_t keylen,
    char** errptr) {
  SaveError(errptr, db->rep->Delete(options->rep, Slice(key, keylen)));
}

void rocksdb_flush(
    rocksdb_t* db,
    const rocksdb_flushoptions_t* options,
    char** errptr) {
  SaveError(errptr, db->rep->Flush(options->rep));
}

char* rocksdb_property_value(
    rocksdb_t* db,
    const char* propname) {
  std::string tmp;
  if (db->rep->GetProperty(Slice(propname), &tmp)) {
    // We use strdup() since we expect human readable output.
    return strdup(tmp.c_str());
  } else {
    return nullptr;
  }
}

//////////////////////////// rocksdb_writebatch_t
rocksdb_writebatch_t* rocksdb_writebatch_create() {
  return new rocksdb_writebatch_t;
}

rocksdb_writebatch_t* rocksdb_writebatch_create_from(const char* rep,
                                                     size_t size) {
  rocksdb_writebatch_t* b = new rocksdb_writebatch_t;
  b->rep = WriteBatch(std::string(rep, size));
  return b;
}

void rocksdb_writebatch_destroy(rocksdb_writebatch_t* b) {
  delete b;
}

void rocksdb_writebatch_clear(rocksdb_writebatch_t* b) {
  b->rep.Clear();
}

int rocksdb_writebatch_count(rocksdb_writebatch_t* b) {
  return b->rep.Count();
}

void rocksdb_writebatch_put(
    rocksdb_writebatch_t* b,
    const char* key, size_t klen,
    const char* val, size_t vlen) {
  b->rep.Put(Slice(key, klen), Slice(val, vlen));
}

void rocksdb_writebatch_delete(
    rocksdb_writebatch_t* b,
    const char* key, size_t klen) {
  b->rep.Delete(Slice(key, klen));
}

void rocksdb_write(
    rocksdb_t* db,
    const rocksdb_writeoptions_t* options,
    rocksdb_writebatch_t* batch,
    char** errptr) {
  SaveError(errptr, db->rep->Write(options->rep, &batch->rep));
}

//////////////////////////// rocksdb_options_t
rocksdb_options_t* rocksdb_options_create() {
  return new rocksdb_options_t;
}

void rocksdb_options_set_create_if_missing(
    rocksdb_options_t* opt, unsigned char v) {
  opt->rep.create_if_missing = v;
}

void rocksdb_options_set_block_based_table_factory(
    rocksdb_options_t *opt,
    rocksdb_block_based_table_options_t* table_options) {
  if (table_options) {
    opt->rep.table_factory.reset(
        rocksdb::NewBlockBasedTableFactory(table_options->rep));
  }
}

void rocksdb_options_set_hash_skip_list_rep(
    rocksdb_options_t *opt, size_t bucket_count,
    int32_t skiplist_height, int32_t skiplist_branching_factor) {
  static rocksdb::MemTableRepFactory* factory = 0;
  if (!factory) {
    factory = rocksdb::NewHashSkipListRepFactory(
        bucket_count, skiplist_height, skiplist_branching_factor);
  }
  opt->rep.memtable_factory.reset(factory);
}

void rocksdb_options_set_hash_link_list_rep(
    rocksdb_options_t *opt, size_t bucket_count) {
  static rocksdb::MemTableRepFactory* factory = 0;
  if (!factory) {
    factory = rocksdb::NewHashLinkListRepFactory(bucket_count);
  }
  opt->rep.memtable_factory.reset(factory);
}

void rocksdb_options_set_compression(rocksdb_options_t* opt, int t) {
  opt->rep.compression = static_cast<CompressionType>(t);
}

void rocksdb_options_set_compression_per_level(rocksdb_options_t* opt,
                                               int* level_values,
                                               size_t num_levels) {
  opt->rep.compression_per_level.resize(num_levels);
  for (size_t i = 0; i < num_levels; ++i) {
    opt->rep.compression_per_level[i] =
      static_cast<CompressionType>(level_values[i]);
  }
}

void rocksdb_options_set_min_level_to_compress(rocksdb_options_t* opt, int level) {
  if (level >= 0) {
    assert(level <= opt->rep.num_levels);
    opt->rep.compression_per_level.resize(opt->rep.num_levels);
    for (int i = 0; i < level; i++) {
      opt->rep.compression_per_level[i] = rocksdb::kNoCompression;
    }
    for (int i = level; i < opt->rep.num_levels; i++) {
      opt->rep.compression_per_level[i] = opt->rep.compression;
    }
  }
}

void rocksdb_options_set_plain_table_factory(
    rocksdb_options_t *opt, uint32_t user_key_len, int bloom_bits_per_key,
    double hash_table_ratio, size_t index_sparseness, int encodingType) {
  static rocksdb::TableFactory* factory = 0;
  if (!factory) {
    rocksdb::PlainTableOptions options;
    options.user_key_len = user_key_len;
    options.bloom_bits_per_key = bloom_bits_per_key;
    options.hash_table_ratio = hash_table_ratio;
    options.index_sparseness = index_sparseness;
		options.encoding_type = (rocksdb::EncodingType)encodingType;
		
    factory = rocksdb::NewPlainTableFactory(options);
  }
  opt->rep.table_factory.reset(factory);
}

//////////////////////////// rocksdb_readoptions_t
rocksdb_readoptions_t* rocksdb_readoptions_create() {
  return new rocksdb_readoptions_t;
}

void rocksdb_readoptions_destroy(rocksdb_readoptions_t* opt) {
  delete opt;
}

void rocksdb_readoptions_set_fill_cache(
    rocksdb_readoptions_t* opt, unsigned char v) {
  opt->rep.fill_cache = v;
}

void rocksdb_readoptions_set_total_order_seek(
		rocksdb_readoptions_t* opt, unsigned char v) {
	opt->rep.total_order_seek = v;
}

void rocksdb_readoptions_set_prefix_same_as_start(
		rocksdb_readoptions_t* opt, unsigned char v) {
	opt->rep.prefix_same_as_start = v;
}

//////////////////////////// rocksdb_writeoptions_t
rocksdb_writeoptions_t* rocksdb_writeoptions_create() {
  return new rocksdb_writeoptions_t;
}

void rocksdb_writeoptions_destroy(rocksdb_writeoptions_t* opt) {
  delete opt;
}

void rocksdb_writeoptions_set_sync(
    rocksdb_writeoptions_t* opt, unsigned char v) {
  opt->rep.sync = v;
}

//////////////////////////// rocksdb_flushoptions_t
rocksdb_flushoptions_t* rocksdb_flushoptions_create() {
  return new rocksdb_flushoptions_t;
}

//////////////////////////// rocksdb_iterator_t
rocksdb_iterator_t* rocksdb_create_iterator(
    rocksdb_t* db,
    const rocksdb_readoptions_t* options) {
  rocksdb_iterator_t* result = new rocksdb_iterator_t;
  result->rep = db->rep->NewIterator(options->rep);
  return result;
}

void rocksdb_iter_destroy(rocksdb_iterator_t* iter) {
  delete iter->rep;
  delete iter;
}

unsigned char rocksdb_iter_valid(const rocksdb_iterator_t* iter) {
  return iter->rep->Valid();
}

void rocksdb_iter_seek_to_first(rocksdb_iterator_t* iter) {
  iter->rep->SeekToFirst();
}

void rocksdb_iter_seek_to_last(rocksdb_iterator_t* iter) {
  iter->rep->SeekToLast();
}

void rocksdb_iter_seek(rocksdb_iterator_t* iter, const char* k, size_t klen) {
  iter->rep->Seek(Slice(k, klen));
}

void rocksdb_iter_next(rocksdb_iterator_t* iter) {
  iter->rep->Next();
}

void rocksdb_iter_prev(rocksdb_iterator_t* iter) {
  iter->rep->Prev();
}

const char* rocksdb_iter_key(const rocksdb_iterator_t* iter, size_t* klen) {
  Slice s = iter->rep->key();
  *klen = s.size();
  return s.data();
}

const char* rocksdb_iter_value(const rocksdb_iterator_t* iter, size_t* vlen) {
  Slice s = iter->rep->value();
  *vlen = s.size();
  return s.data();
}

void rocksdb_iter_get_error(const rocksdb_iterator_t* iter, char** errptr) {
  SaveError(errptr, iter->rep->status());
}

//////////////////////////// rocksdb_filterpolicy_t
struct rocksdb_filterpolicy_t : public FilterPolicy {
  void* state_;
  void (*destructor_)(void*);
  const char* (*name_)(void*);
  char* (*create_)(
      void*,
      const char* const* key_array, const size_t* key_length_array,
      int num_keys,
      size_t* filter_length);
  unsigned char (*key_match_)(
      void*,
      const char* key, size_t length,
      const char* filter, size_t filter_length);
  void (*delete_filter_)(
      void*,
      const char* filter, size_t filter_length);

  virtual ~rocksdb_filterpolicy_t() {
    (*destructor_)(state_);
  }

  virtual const char* Name() const override { return (*name_)(state_); }

  virtual void CreateFilter(const Slice* keys, int n,
                            std::string* dst) const override {
    std::vector<const char*> key_pointers(n);
    std::vector<size_t> key_sizes(n);
    for (int i = 0; i < n; i++) {
      key_pointers[i] = keys[i].data();
      key_sizes[i] = keys[i].size();
    }
    size_t len;
    char* filter = (*create_)(state_, &key_pointers[0], &key_sizes[0], n, &len);
    dst->append(filter, len);

    if (delete_filter_ != nullptr) {
      (*delete_filter_)(state_, filter, len);
    } else {
      free(filter);
    }
  }

  virtual bool KeyMayMatch(const Slice& key,
                           const Slice& filter) const override {
    return (*key_match_)(state_, key.data(), key.size(),
                         filter.data(), filter.size());
  }
};

rocksdb_filterpolicy_t* rocksdb_filterpolicy_create(
    void* state,
    void (*destructor)(void*),
    char* (*create_filter)(
        void*,
        const char* const* key_array, const size_t* key_length_array,
        int num_keys,
        size_t* filter_length),
    unsigned char (*key_may_match)(
        void*,
        const char* key, size_t length,
        const char* filter, size_t filter_length),
    void (*delete_filter)(
        void*,
        const char* filter, size_t filter_length),
    const char* (*name)(void*)) {
  rocksdb_filterpolicy_t* result = new rocksdb_filterpolicy_t;
  result->state_ = state;
  result->destructor_ = destructor;
  result->create_ = create_filter;
  result->key_match_ = key_may_match;
  result->delete_filter_ = delete_filter;
  result->name_ = name;
  return result;
}

void rocksdb_filterpolicy_destroy(rocksdb_filterpolicy_t* filter) {
  delete filter;
}

void rdbc_destruct_handler(void* state) { }
void rdbc_filterpolicy_delete_filter(void* state, const char* v, size_t s) { }

rocksdb_filterpolicy_t* rdbc_filterpolicy_create(uintptr_t idx) {
  return rocksdb_filterpolicy_create(
    (void*)idx,
    rdbc_destruct_handler,
    (char* (*)(void*, const char* const*, const size_t*, int, size_t*))(rdbc_filterpolicy_create_filter),
    (unsigned char (*)(void*, const char*, size_t, const char*, size_t))(rdbc_filterpolicy_key_may_match),
    rdbc_filterpolicy_delete_filter,
    (const char *(*)(void*))(rdbc_filterpolicy_name));
}

rocksdb_filterpolicy_t* rocksdb_filterpolicy_create_bloom_format(int bits_per_key, bool original_format) {
  // Make a rocksdb_filterpolicy_t, but override all of its methods so
  // they delegate to a NewBloomFilterPolicy() instead of user
  // supplied C functions.
  struct Wrapper : public rocksdb_filterpolicy_t {
    const FilterPolicy* rep_;
    ~Wrapper() { delete rep_; }
    const char* Name() const override { return rep_->Name(); }
    void CreateFilter(const Slice* keys, int n,
                      std::string* dst) const override {
      return rep_->CreateFilter(keys, n, dst);
    }
    bool KeyMayMatch(const Slice& key, const Slice& filter) const override {
      return rep_->KeyMayMatch(key, filter);
    }
    static void DoNothing(void*) { }
  };
  Wrapper* wrapper = new Wrapper;
  wrapper->rep_ = NewBloomFilterPolicy(bits_per_key, original_format);
  wrapper->state_ = nullptr;
  wrapper->delete_filter_ = nullptr;
  wrapper->destructor_ = &Wrapper::DoNothing;
  return wrapper;
}

rocksdb_filterpolicy_t* rocksdb_filterpolicy_create_bloom_full(int bits_per_key) {
  return rocksdb_filterpolicy_create_bloom_format(bits_per_key, false);
}

rocksdb_filterpolicy_t* rocksdb_filterpolicy_create_bloom(int bits_per_key) {
  return rocksdb_filterpolicy_create_bloom_format(bits_per_key, true);
}

//////////////////////////// rocksdb_cache_t
rocksdb_cache_t* rocksdb_cache_create_lru(size_t capacity) {
  rocksdb_cache_t* c = new rocksdb_cache_t;
  c->rep = NewLRUCache(capacity);
  return c;
}

void rocksdb_cache_destroy(rocksdb_cache_t* cache) {
  delete cache;
}

void rocksdb_cache_set_capacity(rocksdb_cache_t* cache, size_t capacity) {
  cache->rep->SetCapacity(capacity);
}

//////////////////////////// rocksdb_block_based_table_options_t
rocksdb_block_based_table_options_t*
rocksdb_block_based_options_create() {
  return new rocksdb_block_based_table_options_t;
}

void rocksdb_block_based_options_destroy(
    rocksdb_block_based_table_options_t* options) {
  delete options;
}

void rocksdb_block_based_options_set_block_size(
    rocksdb_block_based_table_options_t* options, size_t block_size) {
  options->rep.block_size = block_size;
}

void rocksdb_block_based_options_set_filter_policy(
    rocksdb_block_based_table_options_t* options,
    rocksdb_filterpolicy_t* filter_policy) {
  options->rep.filter_policy.reset(filter_policy);
}

void rocksdb_block_based_options_set_no_block_cache(
    rocksdb_block_based_table_options_t* options,
    unsigned char no_block_cache) {
  options->rep.no_block_cache = no_block_cache;
}

void rocksdb_block_based_options_set_block_cache(
    rocksdb_block_based_table_options_t* options,
    rocksdb_cache_t* block_cache) {
  if (block_cache) {
    options->rep.block_cache = block_cache->rep;
  }
}

void rocksdb_block_based_options_set_block_cache_compressed(
    rocksdb_block_based_table_options_t* options,
    rocksdb_cache_t* block_cache_compressed) {
  if (block_cache_compressed) {
    options->rep.block_cache_compressed = block_cache_compressed->rep;
  }
}

void rocksdb_block_based_options_set_whole_key_filtering(
    rocksdb_block_based_table_options_t* options, unsigned char v) {
  options->rep.whole_key_filtering = v;
}

void rocksdb_block_based_options_set_index_type(
    rocksdb_block_based_table_options_t* options, int v) {
  options->rep.index_type = static_cast<BlockBasedTableOptions::IndexType>(v);
}

//////////////////////////// rocksdb_slicetransform_t
struct rocksdb_slicetransform_t : public SliceTransform {
  void* state_;
  void (*destructor_)(void*);
  const char* (*name_)(void*);
  char* (*transform_)(
      void*,
      const char* key, size_t length,
      size_t* dst_length);
  unsigned char (*in_domain_)(
      void*,
      const char* key, size_t length);
  unsigned char (*in_range_)(
      void*,
      const char* key, size_t length);

  virtual ~rocksdb_slicetransform_t() {
    (*destructor_)(state_);
  }

  virtual const char* Name() const override { return (*name_)(state_); }

  virtual Slice Transform(const Slice& src) const override {
    size_t len;
    char* dst = (*transform_)(state_, src.data(), src.size(), &len);
    return Slice(dst, len);
  }

  virtual bool InDomain(const Slice& src) const override {
    return (*in_domain_)(state_, src.data(), src.size());
  }

  virtual bool InRange(const Slice& src) const override {
    return (*in_range_)(state_, src.data(), src.size());
  }
};

rocksdb_slicetransform_t* rocksdb_slicetransform_create(
    void* state,
    void (*destructor)(void*),
    char* (*transform)(
        void*,
        const char* key, size_t length,
        size_t* dst_length),
    unsigned char (*in_domain)(
        void*,
        const char* key, size_t length),
    unsigned char (*in_range)(
        void*,
        const char* key, size_t length),
    const char* (*name)(void*)) {
  rocksdb_slicetransform_t* result = new rocksdb_slicetransform_t;
  result->state_ = state;
  result->destructor_ = destructor;
  result->transform_ = transform;
  result->in_domain_ = in_domain;
  result->in_range_ = in_range;
  result->name_ = name;
  return result;
}

void rocksdb_slicetransform_destroy(rocksdb_slicetransform_t* st) {
  delete st;
}

rocksdb_slicetransform_t* rocksdb_slicetransform_create_fixed_prefix(size_t prefixLen) {
  struct Wrapper : public rocksdb_slicetransform_t {
    const SliceTransform* rep_;
    ~Wrapper() { delete rep_; }
    const char* Name() const override { return rep_->Name(); }
    Slice Transform(const Slice& src) const override {
      return rep_->Transform(src);
    }
    bool InDomain(const Slice& src) const override {
      return rep_->InDomain(src);
    }
    bool InRange(const Slice& src) const override { return rep_->InRange(src); }
    static void DoNothing(void*) { }
  };
  Wrapper* wrapper = new Wrapper;
  wrapper->rep_ = rocksdb::NewFixedPrefixTransform(prefixLen);
  wrapper->state_ = nullptr;
  wrapper->destructor_ = &Wrapper::DoNothing;
  return wrapper;
}

void rocksdb_options_set_prefix_extractor(
    rocksdb_options_t* opt, rocksdb_slicetransform_t* prefix_extractor) {
  opt->rep.prefix_extractor.reset(prefix_extractor);
}

rocksdb_slicetransform_t* rdbc_slicetransform_create(uintptr_t idx) {
    return rocksdb_slicetransform_create(
    	(void*)idx,
    	rdbc_destruct_handler,
    	(char* (*)(void*, const char*, size_t, size_t*))(rdbc_slicetransform_transform),
    	(unsigned char (*)(void*, const char*, size_t))(rdbc_slicetransform_in_domain),
    	(unsigned char (*)(void*, const char*, size_t))(rdbc_slicetransform_in_range),
    	(const char* (*)(void*))(rdbc_slicetransform_name));
}
