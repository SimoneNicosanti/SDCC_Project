# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: file_transfer/File.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x18\x66ile_transfer/File.proto\x12\x06\x63lient\":\n\x11\x46ileUploadRequest\x12\x12\n\nrequest_id\x18\x01 \x01(\x05\x12\x11\n\tfile_name\x18\x02 \x01(\t\"A\n\tFileChunk\x12\x12\n\nrequest_id\x18\x01 \x01(\x05\x12\x11\n\tfile_name\x18\x02 \x01(\t\x12\r\n\x05\x63hunk\x18\x03 \x01(\x0c\"/\n\x08Response\x12\x12\n\nrequest_id\x18\x01 \x01(\x05\x12\x0f\n\x07success\x18\x02 \x01(\x08\x32z\n\x0b\x46ileService\x12\x38\n\x06Upload\x12\x19.client.FileUploadRequest\x1a\x11.client.FileChunk0\x01\x12\x31\n\x08\x44ownload\x12\x11.client.FileChunk\x1a\x10.client.Response(\x01\x42\nZ\x08../protob\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'file_transfer.File_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z\010../proto'
  _globals['_FILEUPLOADREQUEST']._serialized_start=36
  _globals['_FILEUPLOADREQUEST']._serialized_end=94
  _globals['_FILECHUNK']._serialized_start=96
  _globals['_FILECHUNK']._serialized_end=161
  _globals['_RESPONSE']._serialized_start=163
  _globals['_RESPONSE']._serialized_end=210
  _globals['_FILESERVICE']._serialized_start=212
  _globals['_FILESERVICE']._serialized_end=334
# @@protoc_insertion_point(module_scope)
