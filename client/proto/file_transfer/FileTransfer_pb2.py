# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: file_transfer/FileTransfer.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n file_transfer/FileTransfer.proto\x12\rfile_transfer\":\n\x13\x46ileDownloadRequest\x12\x10\n\x08\x66ileName\x18\x01 \x01(\t\x12\x11\n\trequestId\x18\x02 \x01(\t\"\x1a\n\tFileChunk\x12\r\n\x05\x63hunk\x18\x01 \x01(\x0c\"2\n\x0c\x46ileResponse\x12\x11\n\trequestId\x18\x01 \x01(\t\x12\x0f\n\x07success\x18\x02 \x01(\x08\"8\n\x11\x46ileDeleteRequest\x12\x10\n\x08\x66ileName\x18\x01 \x01(\t\x12\x11\n\trequestId\x18\x02 \x01(\t*\xbd\x01\n\nErrorCodes\x12\x06\n\x02OK\x10\x00\x12\x0c\n\x08S3_ERROR\x10\t\x12\x0f\n\x0b\x43HUNK_ERROR\x10\n\x12\x14\n\x10INVALID_METADATA\x10\x0b\x12\x15\n\x11\x46ILE_CREATE_ERROR\x10\x0c\x12\x18\n\x14\x46ILE_NOT_FOUND_ERROR\x10\r\x12\x14\n\x10\x46ILE_WRITE_ERROR\x10\x0f\x12\x13\n\x0f\x46ILE_READ_ERROR\x10\x10\x12\x16\n\x12STREAM_CLOSE_ERROR\x10\x11\x32\xe5\x01\n\x0b\x46ileService\x12J\n\x08\x44ownload\x12\".file_transfer.FileDownloadRequest\x1a\x18.file_transfer.FileChunk0\x01\x12\x41\n\x06Upload\x12\x18.file_transfer.FileChunk\x1a\x1b.file_transfer.FileResponse(\x01\x12G\n\x06\x44\x65lete\x12 .file_transfer.FileDeleteRequest\x1a\x1b.file_transfer.FileResponse2e\n\x0f\x45\x64geFileService\x12R\n\x10\x44ownloadFromEdge\x12\".file_transfer.FileDownloadRequest\x1a\x18.file_transfer.FileChunk0\x01\x42\x18Z\x16../proto/file_transferb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'file_transfer.FileTransfer_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z\026../proto/file_transfer'
  _globals['_ERRORCODES']._serialized_start=250
  _globals['_ERRORCODES']._serialized_end=439
  _globals['_FILEDOWNLOADREQUEST']._serialized_start=51
  _globals['_FILEDOWNLOADREQUEST']._serialized_end=109
  _globals['_FILECHUNK']._serialized_start=111
  _globals['_FILECHUNK']._serialized_end=137
  _globals['_FILERESPONSE']._serialized_start=139
  _globals['_FILERESPONSE']._serialized_end=189
  _globals['_FILEDELETEREQUEST']._serialized_start=191
  _globals['_FILEDELETEREQUEST']._serialized_end=247
  _globals['_FILESERVICE']._serialized_start=442
  _globals['_FILESERVICE']._serialized_end=671
  _globals['_EDGEFILESERVICE']._serialized_start=673
  _globals['_EDGEFILESERVICE']._serialized_end=774
# @@protoc_insertion_point(module_scope)
