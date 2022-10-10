import json
import os


class DatasetList:
    def __init__(self):
        with open("./configs/datasets.json", "r") as file:
            datasets_file = file.read()
        datasets_file = json.loads(datasets_file)
        self.datasets = datasets_file
    
    
    def json_output(self):
        datasets_json = json.dumps(self.datasets, indent=2)
        return datasets_json