import onnx
model = onnx.load("model/lgb_stock_model.onnx")

for o in model.graph.output:
    print(o.name, o.type.tensor_type.elem_type)