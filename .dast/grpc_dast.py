import grpc
import random
import string
import base64
import logging
from concurrent import futures

import sensor_event_pb2
import sensor_event_pb2_grpc

GRPC_ADDRESS = 'sensor-api:50051'
NUM_EVENTS = 100
FUZZ_STR_LEN = 50
FUZZ_NUM_MAX = 10**6

logging.basicConfig(level=logging.INFO, format='[%(levelname)s] %(message)s')

def random_fuzz_string(length=FUZZ_STR_LEN):
    chars = string.ascii_letters + string.digits + string.punctuation
    return ''.join(random.choice(chars) for _ in range(length))

def random_fuzz_number(max_value=FUZZ_NUM_MAX):
    return random.randint(0, max_value)

def make_fuzz_metric():
    return sensor_event_pb2.Metric(
        snort_timestamp="2025-06-12T00:00:00Z",
        snort_base64_data=base64.b64encode(random_fuzz_string().encode()).decode(),
        snort_dst_address=random_fuzz_string(),
        snort_dst_port=random_fuzz_number(),
        snort_src_address=random_fuzz_string(),
        snort_src_port=random_fuzz_number()
    )

def make_fuzz_event():
    return sensor_event_pb2.SensorEvent(
        metrics=[make_fuzz_metric()],
        event_hash_sha256=''.join(random.choice('abcdef0123456789') for _ in range(64)),
        event_metrics_count=random_fuzz_number(),
        event_seconds=random_fuzz_number(),
        sensor_id=random_fuzz_string(),
        sensor_version=random_fuzz_string(10),
        event_read_at=random_fuzz_number(),
        event_sent_at=random_fuzz_number(),
        event_received_at=random_fuzz_number(),
        snort_interface=random_fuzz_string(8),
        snort_protocol=random_fuzz_string(5),
        snort_message=random_fuzz_string(20),
        snort_priority=random_fuzz_number(10),
        snort_rule_gid=random_fuzz_number(100),
        snort_rule_rev=random_fuzz_number(100),
        snort_rule_sid=random_fuzz_number(1000000),
        snort_rule=random_fuzz_string(15),
        snort_seconds=random_fuzz_number()
    )

def run_dast():
    channel = grpc.insecure_channel(GRPC_ADDRESS)
    stub = sensor_event_pb2_grpc.SensorServiceStub(channel)

    def event_generator():
        for i in range(NUM_EVENTS):
            evt = make_fuzz_event()
            logging.info(f"Sending event #{i+1}")
            yield evt

    try:
        response = stub.StreamData(event_generator())
        logging.info(f"StreamData completed: {response}")
    except grpc.RpcError as e:
        logging.error(f"RPC Error: code={e.code()} details={e.details()}")

if __name__ == '__main__':
    logging.info("Starting gRPC DAST script...")
    run_dast()
