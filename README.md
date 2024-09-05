# VTRGo
VTRGo is an application written in Go for collecting data from ControlLogix PLC's and for displaying the data, sending the data to excel, or sending reports via email.

### Features
- Read and write tags to a ControlLogix PLC
- Configuration using a browser interface
- Simple database storage
- Create simple Logix output files
- Write PLC data to excel
- Capture FactoryTalk alarms in excel
- Send reports to configurable email recipients  

### Configuration

` export EMAIL_CONFIG_FILE="env/email-config.json" `