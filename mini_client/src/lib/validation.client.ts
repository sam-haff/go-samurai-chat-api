export function validateEmail(email: string): boolean {
    let re = /\S+@\S+\.\S+/;
    return re.test(email);
}
export function isAlphanumeric(str: string): boolean {
    let re = /^[A-Za-z0-9]+$/;
    return re.test(str);
}
export function validateUsername(username: string): boolean {
    return username.length > 3 && isAlphanumeric(username);
}
export function validatePassword(pwd: string): boolean {
    return pwd.length > 5;
}