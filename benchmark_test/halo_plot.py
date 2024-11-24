import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

def process_csv(filename, halo_type):
    # Read and filter data
    with open(filename, "r") as f:
        lines = f.readlines()
    filtered_lines = [line.strip() for line in lines if 'BenchmarkGol' in line]

    # Parse data into a list of dictionaries
    data = []
    for line in filtered_lines:
        parts = line.split(',')
        if len(parts) == 3:
            # Extract number of threads and time in nanoseconds
            threads = int(parts[0].split('-')[1])
            time_ns = float(parts[2])  # time in nanoseconds
            data.append({'threads': threads, 'time_ns': time_ns, 'halo_type': halo_type})

    # Create DataFrame from list of dictionaries
    df = pd.DataFrame(data)
    df['time'] = df['time_ns'] / 1e9  # Convert to seconds
    return df

# Process both CSV files
direct_data = process_csv("Direct.csv", "Direct")
broker_data = process_csv("Broker.csv", "Broker")

# Combine datasets
benchmark_data = pd.concat([broker_data,direct_data])

# Create line plot
plt.figure(figsize=(10, 6))
ax = sns.barplot(data=benchmark_data, x='threads', y='time', hue='halo_type')

# Customize plot
ax.set(xlabel='Number of Workers',
       ylabel='Time (seconds)',
       title='Direct vs Broker Halo Performance Comparison')

plt.legend(title='Halo Type')
plt.savefig("Direct_vs_Broker",dpi=300)
plt.show()