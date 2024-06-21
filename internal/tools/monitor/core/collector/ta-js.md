# ta.js

## How to Use?

1. **Make Necessary Updates**:
   - Update the `ta.js` file with the required changes.

2. **Include JavaScript Files**:
   - Run the following command to include JavaScript files and set the namespace to `monitor/metrics-collector`:
    ```shell
    statik -m -src=./ '-include=*.js' -ns=monitor/metrics-collector
    ```

3. **Format the Code**:
   - Use your preferred code formatter to ensure the code is properly formatted.

4. **Update the ChangeLog**:
   - Document the changes made to `ta.js` in the `ChangeLog`.

## ChangeLog

### 2024-06-21

- Fix(server): add local watch.
- Fix(vendor): add empty [] when first init.


### 2024-05-21

- Support for custom attributes in monitoring metrics reporting.
