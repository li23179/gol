import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

def process_csv(filename, system_type):
    # Read and filter data
    contents = open(filename, "r").read().split('\n')
    filtered_lines = [line for line in contents if 'Gol' in line]
    
    # Create temporary file
    temp_file = f"{system_type}.csv"
    with open(temp_file, 'w') as file:
        for line in filtered_lines:
            file.write(line + '\n')
            
    # Read CSV and process
    data = pd.read_csv(temp_file, header=None, names=['name', 'time', 'range'])
    data['time'] /= 1e+9  # Convert to seconds
    data['threads'] = data['name'].str.extract(r'Gol/\w+-(\d+)').apply(pd.to_numeric)
    data['system_type'] = system_type
    return data

# Read both files
direct_data = process_csv("serial_results.csv", "Serial")
broker_data = process_csv("parallel_results.csv", "Parallel")

# Combine datasets
benchmark_data = pd.concat([direct_data, broker_data])

# Create plot
plt.figure(figsize=(10, 6))
ax = sns.barplot(data=benchmark_data, x='threads', y='time', hue='system_type')

# Customize plot
ax.set(xlabel='Number of Workers',
       ylabel='Time (seconds)',
       title='Serial vs Parallel Distributed System Performance Comparison')

plt.legend(title='System Type')
plt.show()
