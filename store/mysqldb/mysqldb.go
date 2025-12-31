package mysqldb

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/gommon/log"
	"github.com/skip2/go-qrcode"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/swissmakers/wireguard-manager/model"
	"github.com/swissmakers/wireguard-manager/util"
)

type MySQLDB struct {
	conn *sql.DB
	dsn  string
}

// New returns a new pointer MySQLDB
func New(dsn string) (*MySQLDB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ans := MySQLDB{
		conn: db,
		dsn:  dsn,
	}
	return &ans, nil
}

func (o *MySQLDB) Init() error {
	// Create tables if they don't exist
	if err := o.createTables(); err != nil {
		return err
	}

	// Initialize default data
	if err := o.initializeDefaultData(); err != nil {
		return err
	}

	return nil
}

func (o *MySQLDB) createTables() error {
	queries := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			username VARCHAR(255) PRIMARY KEY,
			password_hash TEXT NOT NULL,
			admin BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Server interface table
		`CREATE TABLE IF NOT EXISTS server_interface (
			id INT PRIMARY KEY DEFAULT 1,
			addresses JSON NOT NULL,
			listen_port INT NOT NULL,
			post_up TEXT,
			post_down TEXT,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			CHECK (id = 1)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Server keypair table
		`CREATE TABLE IF NOT EXISTS server_keypair (
			id INT PRIMARY KEY DEFAULT 1,
			private_key TEXT NOT NULL,
			public_key TEXT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			CHECK (id = 1)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Global settings table
		`CREATE TABLE IF NOT EXISTS global_settings (
			id INT PRIMARY KEY DEFAULT 1,
			endpoint_address VARCHAR(255) NOT NULL,
			dns_servers JSON NOT NULL,
			mtu INT NOT NULL,
			persistent_keepalive INT NOT NULL,
			firewall_mark VARCHAR(50),
			table_name VARCHAR(50),
			config_file_path VARCHAR(255) NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			CHECK (id = 1)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Hashes table
		`CREATE TABLE IF NOT EXISTS hashes (
			id INT PRIMARY KEY DEFAULT 1,
			client_hash VARCHAR(255) NOT NULL,
			server_hash VARCHAR(255) NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			CHECK (id = 1)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Clients table
		`CREATE TABLE IF NOT EXISTS clients (
			id VARCHAR(255) PRIMARY KEY,
			private_key TEXT,
			public_key TEXT NOT NULL,
			preshared_key TEXT,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255),
			group_name VARCHAR(255),
			subnet_ranges JSON,
			allocated_ips JSON NOT NULL,
			allowed_ips JSON NOT NULL,
			extra_allowed_ips JSON,
			endpoint VARCHAR(255),
			use_server_dns BOOLEAN NOT NULL DEFAULT TRUE,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// API keys table
		`CREATE TABLE IF NOT EXISTS api_keys (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			key_value TEXT NOT NULL,
			permissions JSON NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// API access logs table
		`CREATE TABLE IF NOT EXISTS api_access_logs (
			id VARCHAR(255) PRIMARY KEY,
			api_key_id VARCHAR(255) NOT NULL,
			api_key_name VARCHAR(255) NOT NULL,
			endpoint VARCHAR(255) NOT NULL,
			method VARCHAR(10) NOT NULL,
			ip_address VARCHAR(45) NOT NULL,
			user_agent TEXT,
			status_code INT NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_api_key_id (api_key_id),
			INDEX idx_timestamp (timestamp)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, query := range queries {
		if _, err := o.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	return nil
}

func (o *MySQLDB) initializeDefaultData() error {
	// Initialize server interface
	var count int
	err := o.conn.QueryRow("SELECT COUNT(*) FROM server_interface").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		serverInterface := model.ServerInterface{
			Addresses:  util.LookupEnvOrStrings(util.ServerAddressesEnvVar, []string{util.DefaultServerAddress}),
			ListenPort: util.LookupEnvOrInt(util.ServerListenPortEnvVar, util.DefaultServerPort),
			PostUp:     util.LookupEnvOrString(util.ServerPostUpScriptEnvVar, ""),
			PostDown:   util.LookupEnvOrString(util.ServerPostDownScriptEnvVar, ""),
			UpdatedAt:  time.Now().UTC(),
		}

		addressesJSON, _ := json.Marshal(serverInterface.Addresses)
		_, err = o.conn.Exec(
			"INSERT INTO server_interface (id, addresses, listen_port, post_up, post_down, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			1, addressesJSON, serverInterface.ListenPort, serverInterface.PostUp, serverInterface.PostDown, serverInterface.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	// Initialize server keypair
	err = o.conn.QueryRow("SELECT COUNT(*) FROM server_keypair").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return err
		}

		serverKeyPair := model.ServerKeypair{
			PrivateKey: key.String(),
			PublicKey:  key.PublicKey().String(),
			UpdatedAt:  time.Now().UTC(),
		}

		_, err = o.conn.Exec(
			"INSERT INTO server_keypair (id, private_key, public_key, updated_at) VALUES (?, ?, ?, ?)",
			1, serverKeyPair.PrivateKey, serverKeyPair.PublicKey, serverKeyPair.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	// Initialize global settings
	err = o.conn.QueryRow("SELECT COUNT(*) FROM global_settings").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		endpointAddress := util.LookupEnvOrString(util.EndpointAddressEnvVar, "")
		if endpointAddress == "" {
			publicInterface, err := util.GetPublicIP()
			if err != nil {
				return err
			}
			endpointAddress = publicInterface.IPAddress
		}

		globalSetting := model.GlobalSetting{
			EndpointAddress:     endpointAddress,
			DNSServers:          util.LookupEnvOrStrings(util.DNSEnvVar, []string{util.DefaultDNS}),
			MTU:                 util.LookupEnvOrInt(util.MTUEnvVar, util.DefaultMTU),
			PersistentKeepalive: util.LookupEnvOrInt(util.PersistentKeepaliveEnvVar, util.DefaultPersistentKeepalive),
			FirewallMark:        util.LookupEnvOrString(util.FirewallMarkEnvVar, util.DefaultFirewallMark),
			Table:               util.LookupEnvOrString(util.TableEnvVar, util.DefaultTable),
			ConfigFilePath:      util.LookupEnvOrString(util.ConfigFilePathEnvVar, util.DefaultConfigFilePath),
			UpdatedAt:           time.Now().UTC(),
		}

		dnsJSON, _ := json.Marshal(globalSetting.DNSServers)
		_, err = o.conn.Exec(
			"INSERT INTO global_settings (id, endpoint_address, dns_servers, mtu, persistent_keepalive, firewall_mark, table_name, config_file_path, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			1, globalSetting.EndpointAddress, dnsJSON, globalSetting.MTU, globalSetting.PersistentKeepalive,
			globalSetting.FirewallMark, globalSetting.Table, globalSetting.ConfigFilePath, globalSetting.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	// Initialize hashes
	err = o.conn.QueryRow("SELECT COUNT(*) FROM hashes").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = o.conn.Exec(
			"INSERT INTO hashes (id, client_hash, server_hash) VALUES (?, ?, ?)",
			1, "none", "none",
		)
		if err != nil {
			return err
		}
	}

	// Initialize default user
	err = o.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		user := model.User{
			Username: util.LookupEnvOrString(util.UsernameEnvVar, util.DefaultUsername),
			Admin:    util.DefaultIsAdmin,
		}

		var plaintext string
		var isGeneratedPassword bool

		user.PasswordHash = util.LookupEnvOrString(util.PasswordHashEnvVar, "")
		if user.PasswordHash == "" {
			user.PasswordHash = util.LookupEnvOrFile(util.PasswordHashFileEnvVar, "")
			if user.PasswordHash == "" {
				plaintext = util.LookupEnvOrString(util.PasswordEnvVar, "")
				if plaintext == "" {
					plaintext = util.LookupEnvOrFile(util.PasswordFileEnvVar, "")
					if plaintext == "" {
						// Generate random 8-character password
						plaintext = util.GenerateRandomPassword(8)
						isGeneratedPassword = true
					}
				}
				hash, err := util.HashPassword(plaintext)
				if err != nil {
					return err
				}
				user.PasswordHash = hash
			}
		}

		_, err = o.conn.Exec(
			"INSERT INTO users (username, password_hash, admin) VALUES (?, ?, ?)",
			user.Username, user.PasswordHash, user.Admin,
		)
		if err != nil {
			return err
		}

		util.DBUsersToCRC32[user.Username] = util.GetDBUserCRC32(user)

		// Log and save generated password if it was generated
		if isGeneratedPassword {
			log.Infof("=== INITIAL ADMIN PASSWORD CREATED ===")
			log.Infof("Username: %s", user.Username)
			log.Infof("Password: %s", plaintext)

			// Write password to log file
			logFilePath := path.Join("db", "initial_admin_password.log")
			logContent := fmt.Sprintf("=== INITIAL ADMIN PASSWORD CREATED ===\n")
			logContent += fmt.Sprintf("Username: %s\n", user.Username)
			logContent += fmt.Sprintf("Password: %s\n", plaintext)
			logContent += fmt.Sprintf("Created at: %s\n", time.Now().Format(time.RFC3339))
			logContent += fmt.Sprintf("This password has been saved to: %s\n", logFilePath)
			logContent += "Please change this password after your first login!\n"
			logContent += "==========================================\n"

			if err := os.WriteFile(logFilePath, []byte(logContent), 0600); err != nil {
				log.Warnf("Failed to write password to log file: %v", err)
			} else {
				log.Infof("This password has been saved to: %s", logFilePath)
			}

			log.Infof("Please change this password after your first login!")
			log.Infof("==========================================")
		}
	} else {
		// Load existing users into cache
		users, err := o.GetUsers()
		if err != nil {
			return err
		}
		for _, user := range users {
			util.DBUsersToCRC32[user.Username] = util.GetDBUserCRC32(user)
		}
	}

	return nil
}

// GetUsers returns all users from the database
func (o *MySQLDB) GetUsers() ([]model.User, error) {
	var users []model.User

	rows, err := o.conn.Query("SELECT username, password_hash, admin FROM users")
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.Username, &user.PasswordHash, &user.Admin); err != nil {
			return users, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetUserByName returns a specific user by username
func (o *MySQLDB) GetUserByName(username string) (model.User, error) {
	user := model.User{}
	err := o.conn.QueryRow(
		"SELECT username, password_hash, admin FROM users WHERE username = ?",
		username,
	).Scan(&user.Username, &user.PasswordHash, &user.Admin)

	if err == sql.ErrNoRows {
		return user, fmt.Errorf("user not found")
	}
	return user, err
}

// SaveUser saves a user to the database
func (o *MySQLDB) SaveUser(user model.User) error {
	_, err := o.conn.Exec(
		`INSERT INTO users (username, password_hash, admin) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE password_hash = ?, admin = ?`,
		user.Username, user.PasswordHash, user.Admin,
		user.PasswordHash, user.Admin,
	)
	if err != nil {
		return err
	}

	util.DBUsersToCRC32[user.Username] = util.GetDBUserCRC32(user)
	return nil
}

// DeleteUser deletes a user from the database
func (o *MySQLDB) DeleteUser(username string) error {
	_, err := o.conn.Exec("DELETE FROM users WHERE username = ?", username)
	if err != nil {
		return err
	}
	delete(util.DBUsersToCRC32, username)
	return nil
}

// GetGlobalSettings returns global settings from the database
func (o *MySQLDB) GetGlobalSettings() (model.GlobalSetting, error) {
	settings := model.GlobalSetting{}
	var dnsJSON []byte

	err := o.conn.QueryRow(
		`SELECT endpoint_address, dns_servers, mtu, persistent_keepalive, 
		 firewall_mark, table_name, config_file_path, updated_at 
		 FROM global_settings WHERE id = 1`,
	).Scan(
		&settings.EndpointAddress, &dnsJSON, &settings.MTU, &settings.PersistentKeepalive,
		&settings.FirewallMark, &settings.Table, &settings.ConfigFilePath, &settings.UpdatedAt,
	)

	if err != nil {
		return settings, err
	}

	if err := json.Unmarshal(dnsJSON, &settings.DNSServers); err != nil {
		return settings, err
	}

	return settings, nil
}

// GetServer returns server configuration from the database
func (o *MySQLDB) GetServer() (model.Server, error) {
	server := model.Server{}

	// Get server interface
	serverInterface := model.ServerInterface{}
	var addressesJSON []byte

	err := o.conn.QueryRow(
		"SELECT addresses, listen_port, post_up, post_down, updated_at FROM server_interface WHERE id = 1",
	).Scan(&addressesJSON, &serverInterface.ListenPort, &serverInterface.PostUp, &serverInterface.PostDown, &serverInterface.UpdatedAt)

	if err != nil {
		return server, err
	}

	if err := json.Unmarshal(addressesJSON, &serverInterface.Addresses); err != nil {
		return server, err
	}

	// Get server keypair
	serverKeyPair := model.ServerKeypair{}
	err = o.conn.QueryRow(
		"SELECT private_key, public_key, updated_at FROM server_keypair WHERE id = 1",
	).Scan(&serverKeyPair.PrivateKey, &serverKeyPair.PublicKey, &serverKeyPair.UpdatedAt)

	if err != nil {
		return server, err
	}

	server.Interface = &serverInterface
	server.KeyPair = &serverKeyPair
	return server, nil
}

// SaveServerInterface saves server interface configuration
func (o *MySQLDB) SaveServerInterface(serverInterface model.ServerInterface) error {
	addressesJSON, err := json.Marshal(serverInterface.Addresses)
	if err != nil {
		return err
	}

	serverInterface.UpdatedAt = time.Now().UTC()
	_, err = o.conn.Exec(
		`INSERT INTO server_interface (id, addresses, listen_port, post_up, post_down, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE addresses = ?, listen_port = ?, post_up = ?, post_down = ?, updated_at = ?`,
		1, addressesJSON, serverInterface.ListenPort, serverInterface.PostUp, serverInterface.PostDown, serverInterface.UpdatedAt,
		addressesJSON, serverInterface.ListenPort, serverInterface.PostUp, serverInterface.PostDown, serverInterface.UpdatedAt,
	)

	return err
}

// SaveServerKeyPair saves server keypair
func (o *MySQLDB) SaveServerKeyPair(serverKeyPair model.ServerKeypair) error {
	serverKeyPair.UpdatedAt = time.Now().UTC()
	_, err := o.conn.Exec(
		`INSERT INTO server_keypair (id, private_key, public_key, updated_at) 
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE private_key = ?, public_key = ?, updated_at = ?`,
		1, serverKeyPair.PrivateKey, serverKeyPair.PublicKey, serverKeyPair.UpdatedAt,
		serverKeyPair.PrivateKey, serverKeyPair.PublicKey, serverKeyPair.UpdatedAt,
	)

	return err
}

// SaveGlobalSettings saves global settings
func (o *MySQLDB) SaveGlobalSettings(globalSettings model.GlobalSetting) error {
	dnsJSON, err := json.Marshal(globalSettings.DNSServers)
	if err != nil {
		return err
	}

	globalSettings.UpdatedAt = time.Now().UTC()
	_, err = o.conn.Exec(
		`INSERT INTO global_settings (id, endpoint_address, dns_servers, mtu, persistent_keepalive, 
		 firewall_mark, table_name, config_file_path, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE endpoint_address = ?, dns_servers = ?, mtu = ?, persistent_keepalive = ?, 
		 firewall_mark = ?, table_name = ?, config_file_path = ?, updated_at = ?`,
		1, globalSettings.EndpointAddress, dnsJSON, globalSettings.MTU, globalSettings.PersistentKeepalive,
		globalSettings.FirewallMark, globalSettings.Table, globalSettings.ConfigFilePath, globalSettings.UpdatedAt,
		globalSettings.EndpointAddress, dnsJSON, globalSettings.MTU, globalSettings.PersistentKeepalive,
		globalSettings.FirewallMark, globalSettings.Table, globalSettings.ConfigFilePath, globalSettings.UpdatedAt,
	)

	return err
}

// GetClients returns all clients from the database
func (o *MySQLDB) GetClients(hasQRCode bool) ([]model.ClientData, error) {
	var clients []model.ClientData

	rows, err := o.conn.Query(`
		SELECT id, private_key, public_key, preshared_key, name, email, group_name,
		subnet_ranges, allocated_ips, allowed_ips, extra_allowed_ips, endpoint,
		use_server_dns, enabled, created_at, updated_at
		FROM clients
	`)
	if err != nil {
		return clients, err
	}
	defer rows.Close()

	for rows.Next() {
		client := model.Client{}
		var subnetRangesJSON, allocatedIPsJSON, allowedIPsJSON, extraAllowedIPsJSON []byte
		var privateKey, presharedKey, email, groupName, endpoint sql.NullString

		err := rows.Scan(
			&client.ID, &privateKey, &client.PublicKey, &presharedKey, &client.Name,
			&email, &groupName, &subnetRangesJSON, &allocatedIPsJSON, &allowedIPsJSON,
			&extraAllowedIPsJSON, &endpoint, &client.UseServerDNS, &client.Enabled,
			&client.CreatedAt, &client.UpdatedAt,
		)
		if err != nil {
			return clients, err
		}

		if privateKey.Valid {
			client.PrivateKey = privateKey.String
		}
		if presharedKey.Valid {
			client.PresharedKey = presharedKey.String
		}
		if email.Valid {
			client.Email = email.String
		}
		if groupName.Valid {
			client.Group = groupName.String
		}
		if endpoint.Valid {
			client.Endpoint = endpoint.String
		}

		if subnetRangesJSON != nil {
			if err := json.Unmarshal(subnetRangesJSON, &client.SubnetRanges); err != nil {
				return clients, fmt.Errorf("failed to unmarshal subnet ranges: %v", err)
			}
		}
		if err := json.Unmarshal(allocatedIPsJSON, &client.AllocatedIPs); err != nil {
			return clients, fmt.Errorf("failed to unmarshal allocated IPs: %v", err)
		}
		if err := json.Unmarshal(allowedIPsJSON, &client.AllowedIPs); err != nil {
			return clients, fmt.Errorf("failed to unmarshal allowed IPs: %v", err)
		}
		if extraAllowedIPsJSON != nil {
			if err := json.Unmarshal(extraAllowedIPsJSON, &client.ExtraAllowedIPs); err != nil {
				return clients, fmt.Errorf("failed to unmarshal extra allowed IPs: %v", err)
			}
		}

		clientData := model.ClientData{Client: &client}

		// Generate QR code if requested
		if hasQRCode && client.PrivateKey != "" {
			server, _ := o.GetServer()
			globalSettings, _ := o.GetGlobalSettings()

			png, err := qrcode.Encode(util.BuildClientConfig(client, server, globalSettings), qrcode.Medium, 256)
			if err == nil {
				clientData.QRCode = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
			}
		}

		clients = append(clients, clientData)
	}

	return clients, rows.Err()
}

// GetClientByID returns a specific client by ID
func (o *MySQLDB) GetClientByID(clientID string, qrCodeSettings model.QRCodeSettings) (model.ClientData, error) {
	client := model.Client{}
	clientData := model.ClientData{}

	var subnetRangesJSON, allocatedIPsJSON, allowedIPsJSON, extraAllowedIPsJSON []byte
	var privateKey, presharedKey, email, groupName, endpoint sql.NullString

	err := o.conn.QueryRow(`
		SELECT id, private_key, public_key, preshared_key, name, email, group_name,
		subnet_ranges, allocated_ips, allowed_ips, extra_allowed_ips, endpoint,
		use_server_dns, enabled, created_at, updated_at
		FROM clients WHERE id = ?
	`, clientID).Scan(
		&client.ID, &privateKey, &client.PublicKey, &presharedKey, &client.Name,
		&email, &groupName, &subnetRangesJSON, &allocatedIPsJSON, &allowedIPsJSON,
		&extraAllowedIPsJSON, &endpoint, &client.UseServerDNS, &client.Enabled,
		&client.CreatedAt, &client.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return clientData, fmt.Errorf("client not found")
	}
	if err != nil {
		return clientData, err
	}

	if privateKey.Valid {
		client.PrivateKey = privateKey.String
	}
	if presharedKey.Valid {
		client.PresharedKey = presharedKey.String
	}
	if email.Valid {
		client.Email = email.String
	}
	if groupName.Valid {
		client.Group = groupName.String
	}
	if endpoint.Valid {
		client.Endpoint = endpoint.String
	}

	if subnetRangesJSON != nil {
		if err := json.Unmarshal(subnetRangesJSON, &client.SubnetRanges); err != nil {
			return clientData, fmt.Errorf("failed to unmarshal subnet ranges: %v", err)
		}
	}
	if err := json.Unmarshal(allocatedIPsJSON, &client.AllocatedIPs); err != nil {
		return clientData, fmt.Errorf("failed to unmarshal allocated IPs: %v", err)
	}
	if err := json.Unmarshal(allowedIPsJSON, &client.AllowedIPs); err != nil {
		return clientData, fmt.Errorf("failed to unmarshal allowed IPs: %v", err)
	}
	if extraAllowedIPsJSON != nil {
		if err := json.Unmarshal(extraAllowedIPsJSON, &client.ExtraAllowedIPs); err != nil {
			return clientData, fmt.Errorf("failed to unmarshal extra allowed IPs: %v", err)
		}
	}

	clientData.Client = &client

	// Generate QR code if requested
	if qrCodeSettings.Enabled && client.PrivateKey != "" {
		server, _ := o.GetServer()
		globalSettings, _ := o.GetGlobalSettings()

		if !qrCodeSettings.IncludeDNS {
			globalSettings.DNSServers = []string{}
		}
		if !qrCodeSettings.IncludeMTU {
			globalSettings.MTU = 0
		}

		png, err := qrcode.Encode(util.BuildClientConfig(client, server, globalSettings), qrcode.Medium, 256)
		if err == nil {
			clientData.QRCode = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
		}
	}

	return clientData, nil
}

// SaveClient saves a client to the database
func (o *MySQLDB) SaveClient(client model.Client) error {
	subnetRangesJSON, _ := json.Marshal(client.SubnetRanges)
	allocatedIPsJSON, err := json.Marshal(client.AllocatedIPs)
	if err != nil {
		return err
	}
	allowedIPsJSON, err := json.Marshal(client.AllowedIPs)
	if err != nil {
		return err
	}
	extraAllowedIPsJSON, _ := json.Marshal(client.ExtraAllowedIPs)

	// Use NULL for empty strings
	var privateKey, presharedKey, email, groupName, endpoint interface{}
	if client.PrivateKey != "" {
		privateKey = client.PrivateKey
	}
	if client.PresharedKey != "" {
		presharedKey = client.PresharedKey
	}
	if client.Email != "" {
		email = client.Email
	}
	if client.Group != "" {
		groupName = client.Group
	}
	if client.Endpoint != "" {
		endpoint = client.Endpoint
	}

	_, err = o.conn.Exec(`
		INSERT INTO clients (id, private_key, public_key, preshared_key, name, email, group_name,
		subnet_ranges, allocated_ips, allowed_ips, extra_allowed_ips, endpoint,
		use_server_dns, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		private_key = ?, public_key = ?, preshared_key = ?, name = ?, email = ?, group_name = ?,
		subnet_ranges = ?, allocated_ips = ?, allowed_ips = ?, extra_allowed_ips = ?, endpoint = ?,
		use_server_dns = ?, enabled = ?, updated_at = ?
	`,
		client.ID, privateKey, client.PublicKey, presharedKey, client.Name, email, groupName,
		subnetRangesJSON, allocatedIPsJSON, allowedIPsJSON, extraAllowedIPsJSON, endpoint,
		client.UseServerDNS, client.Enabled, client.CreatedAt, client.UpdatedAt,
		privateKey, client.PublicKey, presharedKey, client.Name, email, groupName,
		subnetRangesJSON, allocatedIPsJSON, allowedIPsJSON, extraAllowedIPsJSON, endpoint,
		client.UseServerDNS, client.Enabled, time.Now().UTC(),
	)

	return err
}

// DeleteClient deletes a client from the database
func (o *MySQLDB) DeleteClient(clientID string) error {
	_, err := o.conn.Exec("DELETE FROM clients WHERE id = ?", clientID)
	return err
}

// GetPath returns an empty string for MySQL (not file-based)
func (o *MySQLDB) GetPath() string {
	return ""
}

// GetHashes returns configuration hashes
func (o *MySQLDB) GetHashes() (model.ClientServerHashes, error) {
	hashes := model.ClientServerHashes{}
	err := o.conn.QueryRow(
		"SELECT client_hash, server_hash FROM hashes WHERE id = 1",
	).Scan(&hashes.Client, &hashes.Server)

	return hashes, err
}

// SaveHashes saves configuration hashes
func (o *MySQLDB) SaveHashes(hashes model.ClientServerHashes) error {
	_, err := o.conn.Exec(
		`INSERT INTO hashes (id, client_hash, server_hash) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE client_hash = ?, server_hash = ?`,
		1, hashes.Client, hashes.Server,
		hashes.Client, hashes.Server,
	)
	return err
}

// GetAPIKeys returns all API keys from the database
func (o *MySQLDB) GetAPIKeys() ([]model.APIKey, error) {
	var keys []model.APIKey

	rows, err := o.conn.Query(`
		SELECT id, name, key_value, permissions, enabled, created_at, updated_at
		FROM api_keys
	`)
	if err != nil {
		return keys, err
	}
	defer rows.Close()

	for rows.Next() {
		key := model.APIKey{}
		var permissionsJSON []byte

		err := rows.Scan(&key.ID, &key.Name, &key.Key, &permissionsJSON, &key.Enabled, &key.CreatedAt, &key.UpdatedAt)
		if err != nil {
			return keys, err
		}

		if err := json.Unmarshal(permissionsJSON, &key.Permissions); err != nil {
			return keys, err
		}

		keys = append(keys, key)
	}

	return keys, rows.Err()
}

// GetAPIKeyByID returns a specific API key by ID
func (o *MySQLDB) GetAPIKeyByID(keyID string) (model.APIKey, error) {
	key := model.APIKey{}
	var permissionsJSON []byte

	err := o.conn.QueryRow(`
		SELECT id, name, key_value, permissions, enabled, created_at, updated_at
		FROM api_keys WHERE id = ?
	`, keyID).Scan(&key.ID, &key.Name, &key.Key, &permissionsJSON, &key.Enabled, &key.CreatedAt, &key.UpdatedAt)

	if err == sql.ErrNoRows {
		return key, fmt.Errorf("API key not found")
	}
	if err != nil {
		return key, err
	}

	if err := json.Unmarshal(permissionsJSON, &key.Permissions); err != nil {
		return key, err
	}

	return key, nil
}

// GetAPIKeyByKey returns a specific API key by the key value
func (o *MySQLDB) GetAPIKeyByKey(keyValue string) (model.APIKey, error) {
	key := model.APIKey{}
	var permissionsJSON []byte

	err := o.conn.QueryRow(`
		SELECT id, name, key_value, permissions, enabled, created_at, updated_at
		FROM api_keys WHERE key_value = ?
	`, keyValue).Scan(&key.ID, &key.Name, &key.Key, &permissionsJSON, &key.Enabled, &key.CreatedAt, &key.UpdatedAt)

	if err == sql.ErrNoRows {
		return key, fmt.Errorf("API key not found")
	}
	if err != nil {
		return key, err
	}

	if err := json.Unmarshal(permissionsJSON, &key.Permissions); err != nil {
		return key, err
	}

	return key, nil
}

// SaveAPIKey saves an API key to the database
func (o *MySQLDB) SaveAPIKey(key model.APIKey) error {
	permissionsJSON, err := json.Marshal(key.Permissions)
	if err != nil {
		return err
	}

	_, err = o.conn.Exec(`
		INSERT INTO api_keys (id, name, key_value, permissions, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		name = ?, key_value = ?, permissions = ?, enabled = ?, updated_at = ?
	`,
		key.ID, key.Name, key.Key, permissionsJSON, key.Enabled, key.CreatedAt, key.UpdatedAt,
		key.Name, key.Key, permissionsJSON, key.Enabled, time.Now().UTC(),
	)

	return err
}

// DeleteAPIKey deletes an API key from the database
func (o *MySQLDB) DeleteAPIKey(keyID string) error {
	_, err := o.conn.Exec("DELETE FROM api_keys WHERE id = ?", keyID)
	return err
}

// SaveAPIAccessLog saves an API access log entry to the database
func (o *MySQLDB) SaveAPIAccessLog(log model.APIAccessLog) error {
	_, err := o.conn.Exec(`
		INSERT INTO api_access_logs (id, api_key_id, api_key_name, endpoint, method, ip_address, user_agent, status_code, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		log.ID, log.APIKeyID, log.APIKeyName, log.Endpoint, log.Method, log.IPAddress, log.UserAgent, log.StatusCode, log.Timestamp,
	)

	return err
}

// GetAPIAccessLogs returns the most recent API access logs
func (o *MySQLDB) GetAPIAccessLogs(limit int) ([]model.APIAccessLog, error) {
	var logs []model.APIAccessLog

	query := "SELECT id, api_key_id, api_key_name, endpoint, method, ip_address, user_agent, status_code, timestamp FROM api_access_logs ORDER BY timestamp DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := o.conn.Query(query)
	if err != nil {
		return logs, err
	}
	defer rows.Close()

	for rows.Next() {
		log := model.APIAccessLog{}
		var userAgent sql.NullString

		err := rows.Scan(&log.ID, &log.APIKeyID, &log.APIKeyName, &log.Endpoint, &log.Method, &log.IPAddress, &userAgent, &log.StatusCode, &log.Timestamp)
		if err != nil {
			return logs, err
		}

		if userAgent.Valid {
			log.UserAgent = userAgent.String
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetAPIAccessLogsByKeyID returns API access logs for a specific API key
func (o *MySQLDB) GetAPIAccessLogsByKeyID(keyID string, limit int) ([]model.APIAccessLog, error) {
	var logs []model.APIAccessLog

	query := "SELECT id, api_key_id, api_key_name, endpoint, method, ip_address, user_agent, status_code, timestamp FROM api_access_logs WHERE api_key_id = ? ORDER BY timestamp DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := o.conn.Query(query, keyID)
	if err != nil {
		return logs, err
	}
	defer rows.Close()

	for rows.Next() {
		log := model.APIAccessLog{}
		var userAgent sql.NullString

		err := rows.Scan(&log.ID, &log.APIKeyID, &log.APIKeyName, &log.Endpoint, &log.Method, &log.IPAddress, &userAgent, &log.StatusCode, &log.Timestamp)
		if err != nil {
			return logs, err
		}

		if userAgent.Valid {
			log.UserAgent = userAgent.String
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}
