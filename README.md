# VTRGo

VTRGo is a Go application for collecting, managing, and visualizing data from ControlLogix PLCs. It features a web-based interface for configuration, data visualization, Excel export, and email notifications.

---

## Features

- **PLC Communication:** Read and write tags to ControlLogix PLCs.
- **Web Interface:** Configure tags, alarms, and users via a modern browser UI.
- **Database Storage:** Uses SQLite for tags, configuration, alarms, and users.
- **Excel Export:** Export PLC data and alarms to Excel files.
- **Alarm Monitoring:** Capture FactoryTalk alarms and log them.
- **Email Reports:** Send data and alarm reports to configurable recipients.
- **Metrics Dashboard:** Real-time metrics and charting.
- **Extensible:** Modular design for easy feature addition.

---

## Getting Started

### Prerequisites

- Go 1.22+
- Git
- ControlLogix PLC on your network

### Installation

1. **Clone the repository:**
   ```sh
   git clone https://github.com/yourusername/vtrgo.git
   cd vtrgo
   ```

2. **Install dependencies:**
   ```sh
   go mod tidy
   ```

3. **Set up environment variables:**
   - For email notifications, set the config file path:
     ```sh
     export EMAIL_CONFIG_FILE="env/email-config.json"
     ```

4. **Build and run:**
   ```sh
   go run main.go
   ```
   The server will start on http://localhost:8088.

---

## Usage

- Open your browser and navigate to http://localhost:8088.
- Use the top navigation to manage:
  - **Config Tags:** System configuration variables.
  - **Data Tags:** PLC tags for data collection.
  - **Alarm Tags:** PLC tags for alarm monitoring.
  - **Email Recipients:** Manage users for email notifications.
  - **Metrics:** View real-time data and charts.
  - **Alarms:** Check and export alarm logs.

---

## Project Structure

- `main.go`: Application entry point and HTTP server.
- `plc/`: PLC communication logic (`plc.PLC`).
- `db/`: Database models and access (tags, config, alarms, users).
- `excel/`: Excel export logic.
- `email/`: Email sending and templates.
- `templates/`: HTML templates for the web UI.
- `resources/`: Static files (CSS, JS).
- `data-routine.go`: Data collection goroutine.
- `alarm-routine.go`: Alarm collection goroutine.

---

## Configuration

- **PLC IP Address:** Set in `main.go` when creating the PLC instance.
- **Email:** Configure SMTP settings in `env/email-config.json`:
  ```json
  {
    "SMTPHost": "smtp.example.com",
    "SMTPPort": 587,
    "Username": "your@email.com",
    "Password": "yourpassword"
  }
  ```
- **Database:** SQLite files are created automatically.

---

## Exporting Data

- Data and alarms are exported to the `output_files/` directory as `.xlsx` files.
- Email reports can be sent manually or triggered by PLC events.

---

## Development

- **Run tests:**  
  *(Add tests to the `/db`, `/plc`, etc. directories as needed)*
  ```sh
  go test ./...
  ```
- **Code formatting:**  
  ```sh
  go fmt ./...
  ```

---

## Contributing

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/your-feature`).
3. Commit your changes.
4. Push to your fork and open a Pull Request.

---

## License

MIT License. See [LICENSE](LICENSE) for details.

---

## Acknowledgements

- [gologix](https://github.com/danomagnum/gologix) for PLC communication.
- [excelize](https://github.com/xuri/excelize) for Excel export.
- [gomail](https://github.com/go-gomail/gomail) for email.
- [HTMX](https://htmx.org/) and [Chart.js](https://www.chartjs.org/) for the web UI.

---

## Contact

For questions or support, open an issue or contact the maintainer.