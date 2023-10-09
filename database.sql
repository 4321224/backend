
CREATE TABLE Customers (
    ID INT AUTO_INCREMENT PRIMARY KEY,
    NIK VARCHAR(255),
    Full_name VARCHAR(255),
    Legal_name VARCHAR(255),
    Birth_place VARCHAR(255),
    Birth_date DATE,
    Salary DECIMAL(15,2),
    KTP_image_path VARCHAR(255),
    Selfie_image_path VARCHAR(255)
);


CREATE TABLE Loans (
   ID INT AUTO_INCREMENT PRIMARY KEY,
   Customer_id INT,
   Tenor_months INT,
   Loan_limit DECIMAL(15,2), 
   Status ENUM('Active', 'Closed', 'Defaulted'),
   FOREIGN KEY (Customer_id) REFERENCES Customers(ID)
);


CREATE TABLE Transactions (
  ID INT AUTO_INCREMENT PRIMARY KEY,
  Loan_id INT,
  Contract_number VARCHAR(255),
  OTR_price DECIMAL (15,2),
  Admin_fee DECIMAL (15,2),
  Installment_amount_per_month DECIMAL (15,2), 
  Interest_rate FLOAT ,
  Asset_name_bought_in_this_transaction TEXT ,
  
FOREIGN KEY (Loan_id) REFERENCES Loans(ID)
);

-- Menambahkan data ke dalam table Customers:
INSERT INTO Customers(NIK, Full_name, Legal_name, Birth_place,Birth_date ,Salary ,KTP_image_path ,Selfie_image_path ) 
VALUES ('1234567890', 'Budi', 'Budi Santoso','Jakarta' ,'1990-01-01' ,5000000.00 ,'path/to/ktp/image' ,'path/to/selfie/image');

-- Menambahkan data ke dalam table Loans:
INSERT INTO Loans(Customer_id,Tenor_months ,Loan_limit ,Status ) 
VALUES ((SELECT ID FROM Customers WHERE NIK = '1234567890'),12 ,20000000.00,'Active');

-- Menambahkan data ke dalam table Transactions:
INSERT INTO Transactions(
	Loan_id ,
	Contract_number ,
	OTR_price ,
	Admin_fee ,
	Installment_amount_per_month ,
	Interest_rate ,
	Asset_name_bought_in_this_transaction )
VALUES ((SELECT ID FROM Loans WHERE Customer_id = (SELECT ID FROM Customers WHERE NIK = '1234567890')),'ABC1234' ,15000000.00 ,100000.00 ,(15000000.00+100000)/12 ,(1/12)*100,'Motor ABC');

