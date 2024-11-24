import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

def process_csv(filename, system_type):
    # Read CSV with no headers
    data = pd.read_csv(filename, header=None, names=['benchmark', 'runs', 'time_ns'])
    
    # Extract thread count
    data['threads'] = data['benchmark'].str.extract(r'-(\d+)-16').astype(int)
    
    # Convert time to seconds
    data['time'] = data['time_ns'] / 1e9
    
    data['system_type'] = system_type
    return data

# Read files
direct_data = process_csv("Direct.csv", "Direct")
broker_data = process_csv("Broker.csv", "Broker")

# Combine data
benchmark_data = pd.concat([direct_data, broker_data])

# Create bar plot with adjusted settings
plt.figure(figsize=(8, 5))  # Smaller figure size
ax = sns.barplot(data=benchmark_data, 
                x='threads', 
                y='time', 
                hue='system_type',
                hue_order=['Broker', 'Direct'],  # Reverse order
                width=0.5)  # Thinner bars

plt.title('Performance Comparison: Direct vs Broker')
plt.xlabel('Number of Worker')
plt.ylabel('Time (seconds)')
plt.legend(title='System Type')
plt.tight_layout()
plt.savefig('halo_comparison.png', dpi=300)  # Higher resolution
plt.close()