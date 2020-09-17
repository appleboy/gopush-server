// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var gorush_pb = require('./gorush_pb.js');
var protobuf_ptypes_struct_struct_pb = require('./protobuf/ptypes/struct/struct_pb.js');

function serialize_proto_HealthCheckRequest(arg) {
  if (!(arg instanceof gorush_pb.HealthCheckRequest)) {
    throw new Error('Expected argument of type proto.HealthCheckRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_proto_HealthCheckRequest(buffer_arg) {
  return gorush_pb.HealthCheckRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_proto_HealthCheckResponse(arg) {
  if (!(arg instanceof gorush_pb.HealthCheckResponse)) {
    throw new Error('Expected argument of type proto.HealthCheckResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_proto_HealthCheckResponse(buffer_arg) {
  return gorush_pb.HealthCheckResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_proto_NotificationReply(arg) {
  if (!(arg instanceof gorush_pb.NotificationReply)) {
    throw new Error('Expected argument of type proto.NotificationReply');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_proto_NotificationReply(buffer_arg) {
  return gorush_pb.NotificationReply.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_proto_NotificationRequest(arg) {
  if (!(arg instanceof gorush_pb.NotificationRequest)) {
    throw new Error('Expected argument of type proto.NotificationRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_proto_NotificationRequest(buffer_arg) {
  return gorush_pb.NotificationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


var GorushService = exports.GorushService = {
  send: {
    path: '/proto.Gorush/Send',
    requestStream: false,
    responseStream: false,
    requestType: gorush_pb.NotificationRequest,
    responseType: gorush_pb.NotificationReply,
    requestSerialize: serialize_proto_NotificationRequest,
    requestDeserialize: deserialize_proto_NotificationRequest,
    responseSerialize: serialize_proto_NotificationReply,
    responseDeserialize: deserialize_proto_NotificationReply,
  },
};

exports.GorushClient = grpc.makeGenericClientConstructor(GorushService);
var HealthService = exports.HealthService = {
  check: {
    path: '/proto.Health/Check',
    requestStream: false,
    responseStream: false,
    requestType: gorush_pb.HealthCheckRequest,
    responseType: gorush_pb.HealthCheckResponse,
    requestSerialize: serialize_proto_HealthCheckRequest,
    requestDeserialize: deserialize_proto_HealthCheckRequest,
    responseSerialize: serialize_proto_HealthCheckResponse,
    responseDeserialize: deserialize_proto_HealthCheckResponse,
  },
};

exports.HealthClient = grpc.makeGenericClientConstructor(HealthService);
